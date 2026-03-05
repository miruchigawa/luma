package session

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	_ "modernc.org/sqlite"

	"luma/config"
)

type SQLiteSession struct {
	client      *whatsmeow.Client
	path        string
	loginMethod string
	phoneNumber string
}

func NewSQLiteSession(cfg *config.SessionConfig) (*SQLiteSession, error) {
	return &SQLiteSession{
		path:        cfg.SQLite.Path,
		loginMethod: cfg.LoginMethod,
		phoneNumber: cfg.PhoneNumber,
	}, nil
}

func (s *SQLiteSession) Connect(ctx context.Context) (*whatsmeow.Client, error) {
	dbLog := waLog.Stdout("Database", "WARN", true)

	// Ensure the directory for SQLite DB exists
	dir := filepath.Dir(s.path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create sqlite directory: %w", err)
		}
	}

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", s.path)
	container, err := sqlstore.New(ctx, "sqlite", dsn, dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlite container: %w", err)
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

func (s *SQLiteSession) Disconnect() {
	if s.client != nil {
		s.client.Disconnect()
	}
}

func (s *SQLiteSession) IsConnected() bool {
	return s.client != nil && s.client.IsConnected()
}
