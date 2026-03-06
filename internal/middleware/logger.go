package middleware

import (
	"fmt"
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

// CommandLoggerMiddleware returns a middleware that logs command execution details
// with structured fields and duration.
func CommandLoggerMiddleware(log *zap.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *command.Context) error {
			if !ctx.Msg.IsCommand {
				return next(ctx)
			}

			start := time.Now()

			// Call the full dispatch chain
			err := next(ctx)

			elapsed := time.Since(start)

			senderJid := ctx.Msg.From.String()
			chatJid := ctx.Msg.Chat.String()

			fields := []zap.Field{
				zap.String("command", ctx.Msg.CommandName),
				zap.String("sender_jid", senderJid),
				zap.String("chat_jid", chatJid),
				zap.Bool("is_group", ctx.Msg.IsGroup),
				zap.Int("args_count", len(ctx.Msg.Args)),
				zap.String("duration", formatDuration(elapsed)),
				zap.Float64("duration_ms", float64(elapsed.Microseconds())/1000.0),
			}

			status := command.StatusOk
			var cmdErr error
			if ctx.Result != nil {
				status = ctx.Result.Status
				if ctx.Result.Err != nil {
					cmdErr = ctx.Result.Err
				}
			} else if err != nil {
				status = command.StatusError
				cmdErr = err
			}
			fields = append(fields, zap.String("status", string(status)))

			if status == command.StatusError && cmdErr != nil {
				fields = append(fields, zap.Error(cmdErr))
			}

			switch status {
			case command.StatusOk:
				log.Info("command executed", fields...)
			case command.StatusGuardDenied:
				log.Warn("command blocked by guard", fields...)
			case command.StatusNotFound:
				log.Debug("command not found", fields...)
			case command.StatusError:
				log.Error("command failed", fields...)
			default:
				log.Info("command executed", fields...)
			}

			return err
		}
	}
}

// formatDuration format duration into a human-friendly string.
// Avoids float precision noise by explicitly formatting specific units.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%dµs", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}
