package session

import (
	"context"
	"fmt"

	"luma/config"

	"go.mau.fi/whatsmeow"
)

// Session defines the interface for managing a whatsmeow WhatsApp connection session.
type Session interface {
	// Connect connects to the WhatsApp server and triggers the login/restore flow.
	Connect(ctx context.Context) (*whatsmeow.Client, error)
	// Disconnect safely disconnects the active WhatsApp client.
	Disconnect()
	// IsConnected returns whether the client is currently connected.
	IsConnected() bool
}

// New creates a new session manager based on the provided configuration driver.
//
// Key design decisions:
// - Factory pattern allows easy swappability between session storages (sqlite, postgres).
// - Relying heavily on whatsmeow's built-in sqlstore for session data abstraction.
func New(cfg *config.SessionConfig) (Session, error) {
	switch cfg.Driver {
	case "sqlite":
		return NewSQLiteSession(cfg)
	case "postgres":
		return NewPostgresSession(cfg)
	default:
		return nil, fmt.Errorf("unsupported session driver: %s", cfg.Driver)
	}
}
