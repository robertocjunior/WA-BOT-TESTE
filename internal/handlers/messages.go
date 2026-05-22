package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"whatsapp-bot/internal/instagram"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// Register returns the main event handler for the WhatsApp client.
func Register(client *whatsmeow.Client) whatsmeow.EventHandler {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.IsFromMe {
				return
			}

			go func() {
				var msgText string
				if v.Message.Conversation != nil {
					msgText = v.Message.GetConversation()
				} else if v.Message.ExtendedTextMessage != nil {
					msgText = v.Message.GetExtendedTextMessage().GetText()
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

				if msgText == "" {
					fmt.Printf("Warning: Empty text from %s\n", v.Info.Sender.String())
				}

				fmt.Printf("Message received from %s: %s\n", v.Info.Sender.String(), msgText)

				instaURL := instagram.ExtractURL(msgText)
				if instaURL != "" {
					_ = client.SendChatPresence(context.Background(), v.Info.Chat, types.ChatPresenceComposing, types.ChatPresenceMediaText)
					_, _ = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
						Conversation: proto.String("⏳ Baixando vídeo..."),
					})

					defer func() {
						_ = client.SendChatPresence(context.Background(), v.Info.Chat, types.ChatPresencePaused, types.ChatPresenceMediaText)
					}()

					videoData, err := instagram.FetchVideo(instaURL)
					if err != nil {
						fmt.Printf("Error processing Instagram link: %v\n", err)
						_, _ = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
							Conversation: proto.String("❌ Erro ao processar o link do Instagram."),
						})
						return
					}

					uploadResp, err := client.Upload(context.Background(), videoData, whatsmeow.MediaVideo)
					if err != nil {
						fmt.Printf("Error uploading video: %v\n", err)
						_, _ = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
							Conversation: proto.String("❌ Erro ao enviar o vídeo para o WhatsApp."),
						})
						return
					}

					videoMsg := &waE2E.VideoMessage{
						URL:           proto.String(uploadResp.URL),
						DirectPath:    proto.String(uploadResp.DirectPath),
						MediaKey:      uploadResp.MediaKey,
						Mimetype:      proto.String("video/mp4"),
						FileEncSHA256: uploadResp.FileEncSHA256,
						FileSHA256:    uploadResp.FileSHA256,
						FileLength:    proto.Uint64(uint64(len(videoData))),
					}

					_, err = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{VideoMessage: videoMsg})
					if err != nil {
						fmt.Printf("Error sending video message: %v\n", err)
					}
					return
				}

				// Old "oi" logic
				_ = client.SendChatPresence(context.Background(), v.Info.Chat, types.ChatPresenceComposing, types.ChatPresenceMediaText)
				time.Sleep(1 * time.Second)

				responseText := "oi"
				msgLower := strings.ToLower(msgText)
				if v.Message.ExtendedTextMessage != nil {
					ext := v.Message.GetExtendedTextMessage()
					msgLower += " " + strings.ToLower(ext.GetMatchedText()) + " " + strings.ToLower(ext.GetDescription()) + " " + strings.ToLower(ext.GetTitle())
				}

				if strings.Contains(msgLower, "instagram.com/reel/") || strings.Contains(msgLower, "instagram.com/reels/") || strings.Contains(msgLower, "instagram.com/p/") {
					responseText = "reels: " + msgText
				}

				_, err := client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{Conversation: proto.String(responseText)})
				if err != nil {
					fmt.Printf("Error sending message: %v\n", err)
				}
				_ = client.SendChatPresence(context.Background(), v.Info.Chat, types.ChatPresencePaused, types.ChatPresenceMediaText)
			}()
		}
	}
}
