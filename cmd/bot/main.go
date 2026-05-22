package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"whatsapp-bot/internal/handlers"
	"whatsapp-bot/internal/whatsapp"
)

func main() {
	// Initialize WhatsApp Client
	client, err := whatsapp.InitClient()
	if err != nil {
		fmt.Printf("Error initializing WhatsApp client: %v\n", err)
		os.Exit(1)
	}

	// Register Event Handlers
	client.AddEventHandler(handlers.Register(client))

	fmt.Println("Bot is running. Press Ctrl+C to exit.")

	// Wait for interrupt signal to gracefully shut down
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Shutting down...")
	client.Disconnect()
}
