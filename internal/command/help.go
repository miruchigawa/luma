package command

import (
	"fmt"
	"strings"
)

type HelpCommand struct {
	BaseCommand
	registry *Registry
}

func NewHelpCommand(r *Registry) *HelpCommand {
	return &HelpCommand{registry: r}
}

func (c *HelpCommand) Name() string {
	return "help"
}

func (c *HelpCommand) Description() string {
	return "Lists all registered commands"
}

func (c *HelpCommand) Execute(ctx *Context) error {
	cmds := c.registry.Commands()

	var sb strings.Builder
	sb.WriteString("🤖 *Bot Commands*\n\n")

	for _, cmd := range cmds {
		sb.WriteString(fmt.Sprintf("• *!%s* - %s\n", cmd.Name(), cmd.Description()))
	}

	return ctx.Reply(sb.String())
}
