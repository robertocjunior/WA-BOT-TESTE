package handlers

import (
	"context"
	"strings"
	"time"

	"whatsapp-bot/internal/instagram"

	"github.com/rs/zerolog/log"
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
				// Panic recovery to keep the bot running 24/7
				defer func() {
					if r := recover(); r != nil {
						log.Error().Interface("panic", r).Msg("Recovered from panic in message handler")
					}
				}()

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
					log.Debug().Str("sender", v.Info.Sender.String()).Msg("Empty text message received")
				}

				log.Info().
					Str("sender", v.Info.Sender.String()).
					Str("chat", v.Info.Chat.String()).
					Str("message", msgText).
					Msg("Message received")

				instaURL := instagram.ExtractURL(msgText)
				if instaURL != "" {
					log.Info().Str("url", instaURL).Msg("Instagram URL detected")
					
					_ = client.SendChatPresence(context.Background(), v.Info.Chat, types.ChatPresenceComposing, types.ChatPresenceMediaText)
					_, _ = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
						Conversation: proto.String("⏳ Baixando vídeo..."),
					})

					defer func() {
						_ = client.SendChatPresence(context.Background(), v.Info.Chat, types.ChatPresencePaused, types.ChatPresenceMediaText)
					}()

					videoData, err := instagram.FetchVideo(instaURL)
					if err != nil {
						log.Error().Err(err).Str("url", instaURL).Msg("Failed to process Instagram link")
						_, _ = client.SendMessage(context.Background(), v.Info.Chat, &waE2E.Message{
							Conversation: proto.String("❌ Erro ao processar o link do Instagram."),
						})
						return
					}

					log.Info().Int("size", len(videoData)).Msg("Uploading video to WhatsApp")
					uploadResp, err := client.Upload(context.Background(), videoData, whatsmeow.MediaVideo)
					if err != nil {
						log.Error().Err(err).Msg("Failed to upload video to WhatsApp")
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
						log.Error().Err(err).Msg("Failed to send video message")
					} else {
						log.Info().Msg("Video message sent successfully")
					}
					return
				}

				// Basic response logic
				_ = client.SendChatPresence(context.Background(), v.Info.Chat, types.ChatPresenceComposing, types.ChatPresenceMediaText)
				time.Sleep(500 * time.Millisecond)

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
					log.Error().Err(err).Msg("Failed to send text response")
				}
				_ = client.SendChatPresence(context.Background(), v.Info.Chat, types.ChatPresencePaused, types.ChatPresenceMediaText)
			}()
		case *events.Connected:
			log.Info().Msg("Connected to WhatsApp")
		case *events.Disconnected:
			log.Warn().Msg("Disconnected from WhatsApp")
		}
	}
}
