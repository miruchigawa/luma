package middleware

import (
	"time"

	"go.uber.org/zap"

	"luma/internal/command"
)

// Logger returns a middleware that logs command execution details.
func Logger() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *command.Context) error {
			if !ctx.Msg.IsCommand {
				return next(ctx)
			}

			start := time.Now()

			ctx.Logger.Info("Command received",
				zap.String("command", ctx.Msg.CommandName),
				zap.String("sender", ctx.Msg.From.ToNonAD().String()),
				zap.String("chat", ctx.Msg.Chat.ToNonAD().String()),
				zap.Bool("is_group", ctx.Msg.IsGroup),
			)

			err := next(ctx)

			duration := time.Since(start)

			if err != nil {
				ctx.Logger.Error("Command executed with error",
					zap.String("command", ctx.Msg.CommandName),
					zap.Duration("duration", duration),
					zap.Error(err),
				)
			} else {
				ctx.Logger.Info("Command executed successfully",
					zap.String("command", ctx.Msg.CommandName),
					zap.Duration("duration", duration),
				)
			}

			return err
		}
	}
}
