package session

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"

	"luma/config"
)

type PostgresSession struct {
	client      *whatsmeow.Client
	dsn         string
	loginMethod string
	phoneNumber string
}

func NewPostgresSession(cfg *config.SessionConfig) (*PostgresSession, error) {
	return &PostgresSession{
		dsn:         cfg.Postgres.DSN,
		loginMethod: cfg.LoginMethod,
		phoneNumber: cfg.PhoneNumber,
	}, nil
}

func (s *PostgresSession) Connect(ctx context.Context) (*whatsmeow.Client, error) {
	dbLog := waLog.Stdout("Database", "WARN", true)

	container, err := sqlstore.New(ctx, "pgx", s.dsn, dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres container: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	clientLog := waLog.Stdout("Client", "WARN", true)
	s.client = whatsmeow.NewClient(deviceStore, clientLog)

	if s.client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := s.client.GetQRChannel(context.Background())
		err = s.client.Connect()
		if err != nil {
			return nil, fmt.Errorf("failed to connect to WhatsApp for provisioning: %w", err)
		}

		for evt := range qrChan {
			if evt.Event == "code" {
				if s.loginMethod == "pair" {
					code, err := s.client.PairPhone(context.Background(), s.phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Windows)")
					if err != nil {
						fmt.Printf("Failed to pair phone: %v\n", err)
						return nil, fmt.Errorf("failed to pair phone: %w", err)
					}
					fmt.Printf("\n>>> Pairing code: %s <<<\n\nPlease enter this code on your phone.\n\n", code)
				} else {
					fmt.Println("No existing session found. Please scan the QR code to login:")
					qr, err := qrcode.New(evt.Code, qrcode.Medium)
					if err == nil {
						fmt.Println(qr.ToString(false))
					} else {
						fmt.Printf("Failed to generate QR: %v\n", err)
					}
				}
			} else {
				fmt.Printf("Login event: %s\n", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = s.client.Connect()
		if err != nil {
			return nil, fmt.Errorf("failed to connect to WhatsApp: %w", err)
		}
	}

	return s.client, nil
}

func (s *PostgresSession) Disconnect() {
	if s.client != nil {
		s.client.Disconnect()
	}
}

func (s *PostgresSession) IsConnected() bool {
	return s.client != nil && s.client.IsConnected()
}
