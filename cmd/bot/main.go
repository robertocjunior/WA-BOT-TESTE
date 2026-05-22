package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"whatsapp-bot/internal/handlers"
	"whatsapp-bot/internal/whatsapp"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Configure zerolog for 24/7 logging
	// In production, you might want pure JSON, but ConsoleWriter is good for dev/logs monitoring.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Info().Msg("Starting WhatsApp Bot Pro...")

	// Initialize WhatsApp Client
	client, err := whatsapp.InitClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing WhatsApp client")
	}

	// Register Event Handlers
	client.AddEventHandler(handlers.Register(client))

	log.Info().Msg("Bot is running and monitored. Press Ctrl+C to exit.")

	// Wait for interrupt signal to gracefully shut down
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Info().Msg("Shutting down gracefully...")
	client.Disconnect()
	log.Info().Msg("Bot stopped.")
}
