package main

import (
	"context"
	"log"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"luma/config"
	"luma/internal/bot"
	"luma/internal/command"
	"luma/internal/guard"
	"luma/internal/middleware"
	"luma/internal/session"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Logger
	logger := initLogger(cfg.Log.Level)
	defer logger.Sync()

	logger.Info("Starting WhatsApp Bot...")

	// 3. Initialize Session Manager
	sess, err := session.New(&cfg.Session)
	if err != nil {
		logger.Fatal("Failed to initialize session manager", zap.Error(err))
	}

	// 4. Create Command Registry
	registry := command.NewRegistry()

	// 5. Register Commands
	pingCmd := &command.PingCommand{}
	echoCmd := &command.EchoCommand{}
	helpCmd := command.NewHelpCommand(registry)

	// Suppress "unused variable" warnings if building without actually attaching admin guard directly to a cmd yet
	_ = guard.NewAdminGuard(cfg.Bot.AdminJIDs)
	_ = guard.NewGroupGuard()
	_ = guard.NewPrivateGuard()

	registry.Register(pingCmd, echoCmd, helpCmd)

	// 6. Build Middleware Chain
	// Rate limit config: 5 requests per 10 seconds
	rateLimiter := middleware.NewRateLimiter(5, 10*time.Second)

	handler := middleware.Chain(
		registry.Dispatch,
		rateLimiter.Middleware(),
		middleware.Logger(),
	)

	// 7. Initialize and run Bot
	waBot := bot.New(cfg, sess, registry, handler, logger)

	ctx := context.Background()
	if err := waBot.Start(ctx); err != nil {
		logger.Fatal("Bot terminated with error", zap.Error(err))
	}
}

func initLogger(level string) *zap.Logger {
	var l zapcore.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		l = zapcore.InfoLevel
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(l)
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	logger, _ := cfg.Build()
	return logger
}
