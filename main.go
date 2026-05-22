package main

import (
	"context"
	"fmt"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"time"

	_ "modernc.org/sqlite"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

type APIRequest struct {
	URL string `json:"url"`
}

type APIResponse struct {
	Status   string `json:"status"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

func extractInstagramURL(text string) string {
	re := regexp.MustCompile(`(?i)https?://(www\.)?instagram\.com/(reel|reels|p)/[a-zA-Z0-9_-]+`)
	return re.FindString(text)
}

// eventHandler handles incoming events from the WhatsApp client.
// Specifically, it listens for message events and responds with "oi".
func eventHandler(client *whatsmeow.Client) whatsmeow.EventHandler {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			// Ignore messages sent by the bot itself
			if v.Info.IsFromMe {
				return
			}

			// Run in a goroutine to stay agile and not block the event loop
			go func() {
				// Determine the message text from various possible sources
				var msgText string
				if v.Message.Conversation != nil {
					msgText = v.Message.GetConversation()
				} else if v.Message.ExtendedTextMessage != nil {
					msgText = v.Message.GetExtendedTextMessage().GetText()
					// Fallback to MatchedText if Text is empty (common in some link previews)
					if msgText == "" {
						msgText = v.Message.GetExtendedTextMessage().GetMatchedText()
					}
				} else if v.Message.ImageMessage != nil {
					msgText = v.Message.GetImageMessage().GetCaption()
				} else if v.Message.VideoMessage != nil {
					msgText = v.Message.GetVideoMessage().GetCaption()
				} else if v.Message.DocumentMessage != nil {
					msgText = v.Message.GetDocumentMessage().GetCaption()
				}

				// If it's still empty, log the full structure for deep inspection
				if msgText == "" {
					fmt.Printf("Warning: Empty text. Full message structure: %+v\n", v.Message)
				}

				fmt.Printf("Message received from %s (Chat: %s): %s\n", v.Info.Sender.String(), v.Info.Chat.String(), msgText)

				instaURL := extractInstagramURL(msgText)
				if instaURL != "" {
					// Send feedback: "Downloading..."
					_ = client.SendChatPresence(context.Background(), v.Info.Chat, types.ChatPresenceComposing, types.ChatPresenceMediaText)
					_, _ = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
						Conversation: proto.String("⏳ Baixando vídeo..."),
					})

					defer func() {
						// Stop "typing..." animation
						_ = client.SendChatPresence(context.Background(), v.Info.Chat, types.ChatPresencePaused, types.ChatPresenceMediaText)
					}()

					// 1. Call the API to get download link
					apiReq := APIRequest{URL: instaURL}
					jsonData, _ := json.Marshal(apiReq)

					tr := &http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					}
					httpClient := &http.Client{Transport: tr}

					req, err := http.NewRequest("POST", "https://api.int.rbcj.com.br/", strings.NewReader(string(jsonData)))
					if err != nil {
						fmt.Printf("Error creating request: %v\n", err)
						_, _ = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
							Conversation: proto.String("❌ Erro ao processar o link."),
						})
						return
					}
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Accept", "application/json")

					resp, err := httpClient.Do(req)
					if err != nil {
						fmt.Printf("Error calling API: %v\n", err)
						_, _ = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
							Conversation: proto.String("❌ Erro ao conectar com o servidor de download."),
						})
						return
					}
					defer resp.Body.Close()

					var apiResp APIResponse
					if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
						fmt.Printf("Error decoding API response: %v\n", err)
						_, _ = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
							Conversation: proto.String("❌ Resposta inválida do servidor."),
						})
						return
					}

					if apiResp.URL == "" {
						fmt.Printf("API did not return a URL. Response: %+v\n", apiResp)
						_, _ = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
							Conversation: proto.String("❌ Não foi possível obter o vídeo desse link."),
						})
						return
					}

					// 2. Download the video
					videoResp, err := httpClient.Get(apiResp.URL)
					if err != nil {
						fmt.Printf("Error downloading video: %v\n", err)
						_, _ = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
							Conversation: proto.String("❌ Erro ao baixar o vídeo do servidor."),
						})
						return
					}
					defer videoResp.Body.Close()

					videoData, err := io.ReadAll(videoResp.Body)
					if err != nil {
						fmt.Printf("Error reading video data: %v\n", err)
						return
					}

					// 3. Upload to WhatsApp
					uploadResp, err := client.Upload(context.Background(), videoData, whatsmeow.MediaVideo)
					if err != nil {
						fmt.Printf("Error uploading video to WhatsApp: %v\n", err)
						_, _ = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
							Conversation: proto.String("❌ Erro ao enviar o vídeo para o WhatsApp."),
						})
						return
					}

					// 4. Send video message
					videoMsg := &waE2E.VideoMessage{
						URL:           proto.String(uploadResp.URL),
						DirectPath:    proto.String(uploadResp.DirectPath),
						MediaKey:      uploadResp.MediaKey,
						Mimetype:      proto.String("video/mp4"),
						FileEncSHA256: uploadResp.FileEncSHA256,
						FileSHA256:    uploadResp.FileSHA256,
						FileLength:    proto.Uint64(uint64(len(videoData))),
					}

					response := &waE2E.Message{
						VideoMessage: videoMsg,
					}

					_, err = client.SendMessage(context.Background(), v.Info.Chat, response)
					if err != nil {
						fmt.Printf("Error sending video to %s: %v\n", v.Info.Chat.String(), err)
					}
					return
				}

				// If it's not an Instagram link, keep the old logic (respond with "oi")
				// Show "typing..." animation
				_ = client.SendChatPresence(context.Background(), v.Info.Chat, types.ChatPresenceComposing, types.ChatPresenceMediaText)
				
				// Small delay to make the typing animation visible (and feel more natural)
				time.Sleep(1 * time.Second)

				// Determine response text
				responseText := "oi"
				msgLower := strings.ToLower(msgText)

				// If it's an ExtendedTextMessage, search for the link in other metadata fields as well
				if v.Message.ExtendedTextMessage != nil {
					ext := v.Message.GetExtendedTextMessage()
					msgLower += " " + strings.ToLower(ext.GetMatchedText())
					msgLower += " " + strings.ToLower(ext.GetDescription())
					msgLower += " " + strings.ToLower(ext.GetTitle())
				}

				if strings.Contains(msgLower, "instagram.com/reel/") || 
				   strings.Contains(msgLower, "instagram.com/reels/") || 
				   strings.Contains(msgLower, "instagram.com/p/") {
					responseText = "reels: " + msgText
				}

				// Prepare the response message
				response := &waE2E.Message{
					Conversation: proto.String(responseText),
				}

				// Send the response back to the chat
				_, err := client.SendMessage(context.Background(), v.Info.Chat, response)
				if err != nil {
					fmt.Printf("Error sending message to %s: %v\n", v.Info.Chat.String(), err)
				} else {
					fmt.Printf("Responded to %s with '%s'\n", v.Info.Chat.String(), responseText)
				}

				// Stop "typing..." animation
				_ = client.SendChatPresence(context.Background(), v.Info.Chat, types.ChatPresencePaused, types.ChatPresenceMediaText)
			}()
		}
	}
}

func main() {
	// Setup logging
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Setup the SQLite database for session storage
	// ModernC sqlite needs foreign keys enabled.
	// Adding _pragma=journal_mode(WAL) and _pragma=busy_timeout(5000) to avoid "database is locked" errors.
	dbParams := "examplestore.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	container, err := sqlstore.New(context.Background(), "sqlite", dbParams, dbLog)
	if err != nil {
		panic(err)
	}

	// Get the first device from the store
	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	// Initialize the WhatsApp client
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler(client))

	// Check if the device is already logged in
	if client.Store.ID == nil {
		// Not logged in, get QR code channel
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		// Display QR code in terminal
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("QR code generated. Please scan it with your WhatsApp app.")
			} else {
				fmt.Println("QR channel event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		fmt.Println("Successfully connected to WhatsApp.")
	}

	// Wait for interrupt signal to gracefully shut down
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
