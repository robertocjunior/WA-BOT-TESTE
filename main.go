package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

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
				fmt.Printf("Message received from %s (Chat: %s): %s\n", v.Info.Sender.String(), v.Info.Chat.String(), v.Message.GetConversation())

				// Show "typing..." animation
				_ = client.SendChatPresence(context.Background(), v.Info.Chat, types.ChatPresenceComposing, types.ChatPresenceMediaText)
				
				// Small delay to make the typing animation visible (and feel more natural)
				time.Sleep(1 * time.Second)

				// Determine response text
				msgText := v.Message.GetConversation()
				responseText := "oi"
				if strings.Contains(msgText, "https://www.instagram.com/reel/") {
					responseText = "reels"
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
	container, err := sqlstore.New(context.Background(), "sqlite3", "file:examplestore.db?_foreign_keys=on", dbLog)
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
