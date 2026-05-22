package whatsapp

import (
	"context"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
	"github.com/mdp/qrterminal/v3"
	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// InitClient initializes the database, connects to WhatsApp, and returns the client.
func InitClient() (*whatsmeow.Client, error) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Create data directory if it doesn't exist
	_ = os.MkdirAll("data", 0755)
	dbParams := "data/examplestore.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	container, err := sqlstore.New(context.Background(), "sqlite", dbParams, dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get device store: %w", err)
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			return nil, fmt.Errorf("failed to connect for login: %w", err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				log.Info().Msg("QR code generated. Please scan it with your WhatsApp app.")
			} else {
				log.Debug().Str("event", evt.Event).Msg("QR channel event")
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}
		log.Info().Msg("Successfully connected to WhatsApp.")
	}

	return client, nil
}
