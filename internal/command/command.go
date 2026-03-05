package command

import (
	"context"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"luma/internal/guard"
	"luma/internal/parser"
)

// Context holds everything a command needs to execute.
type Context struct {
	Client *whatsmeow.Client
	Msg    *parser.ParsedMessage
	Logger *zap.Logger
}

// Reply sends a text message back to the chat where the command originated.
func (ctx *Context) Reply(text string) error {
	msg := &waProto.Message{
		Conversation: proto.String(text),
	}
	// Note: using ctx.Msg.Chat because it ensures we reply to the group if in group, or user if private
	_, err := ctx.Client.SendMessage(context.Background(), ctx.Msg.Chat, msg)
	return err
}

// Command defines the interface for all bot commands.
type Command interface {
	Name() string
	Aliases() []string
	Description() string
	Guards() []guard.Guard
	Execute(ctx *Context) error
}

// BaseCommand provides default implementations for optional methods
// so concrete commands only override what they need.
type BaseCommand struct{}

func (b *BaseCommand) Aliases() []string     { return nil }
func (b *BaseCommand) Guards() []guard.Guard { return nil }
