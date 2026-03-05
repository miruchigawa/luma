package command

type EchoCommand struct {
	BaseCommand
}

func (c *EchoCommand) Name() string {
	return "echo"
}

func (c *EchoCommand) Description() string {
	return "Echoes the text back"
}

func (c *EchoCommand) Execute(ctx *Context) error {
	if ctx.Msg.RawArgs == "" {
		return ctx.Reply("Usage: !echo <text>")
	}
	return ctx.Reply(ctx.Msg.RawArgs)
}
