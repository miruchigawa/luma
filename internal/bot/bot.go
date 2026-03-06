package bot

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
	"go.uber.org/zap"

	"luma/config"
	"luma/internal/command"
	"luma/internal/middleware"
	"luma/internal/parser"
	"luma/internal/session"
)

// Bot orchestrates the core WhatsApp event loop and dispatches commands.
type Bot struct {
	cfg      *config.Config
	client   *whatsmeow.Client
	session  session.Session
	registry *command.Registry
	logger   *zap.Logger
	handler  middleware.HandlerFunc
}

// New constructs a new Bot instance.
func New(cfg *config.Config, sess session.Session, reg *command.Registry, handler middleware.HandlerFunc, log *zap.Logger) *Bot {
	return &Bot{
		cfg:      cfg,
		session:  sess,
		registry: reg,
		handler:  handler,
		logger:   log,
	}
}

// Start connects to WhatsApp and blocks until Context is canceled or SIGINT/SIGTERM.
func (b *Bot) Start(ctx context.Context) error {
	client, err := b.session.Connect(ctx)
	if err != nil {
		return fmt.Errorf("session connect failed: %w", err)
	}
	b.client = client

	client.AddEventHandler(b.onEvent)

	signalCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	b.logger.Info("Bot is now running. Press CTRL+C to exit.")
	<-signalCtx.Done()

	b.Stop()
	return nil
}

// Stop gracefully shuts down the active bot session.
func (b *Bot) Stop() {
	if b.session != nil {
		b.session.Disconnect()
	}
	b.logger.Info("Bot stopped gracefully.")
}

func (b *Bot) onEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		// Anti-replay: ignore messages older than 2 minutes to allow for clock drift
		if time.Since(v.Info.Timestamp) > 2*time.Minute {
			return
		}

		// Ignore commands from the bot itself unless configured to listen to them
		if v.Info.IsFromMe && !b.cfg.Bot.ListenSelf {
			return
		}

		pm := parser.Parse(v, b.cfg.Bot.CommandPrefixes)
		if pm == nil {
			return
		}

		cmdCtx := &command.Context{
			Client: b.client,
			Msg:    pm,
			Logger: b.logger,
		}

		// Process commands in a goroutine so slow commands don't block event handler loop
		go func() {
			if err := b.handler(cmdCtx); err != nil {
				b.logger.Warn("Command pipeline returned error", zap.Error(err))
			}
		}()
	}
}
