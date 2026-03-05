package command

import (
	"fmt"
	"time"
)

type PingCommand struct {
	BaseCommand
}

func (c *PingCommand) Name() string {
	return "ping"
}

func (c *PingCommand) Description() string {
	return "Check bot latency"
}

func (c *PingCommand) Execute(ctx *Context) error {
	latency := time.Since(ctx.Msg.Timestamp).Milliseconds()
	reply := fmt.Sprintf("🏓 Pong! Latency: %dms", latency)
	return ctx.Reply(reply)
}
