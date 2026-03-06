package command

import (
	"strings"

	"go.uber.org/zap"
)

// Registry manages all available commands and handles dispatching.
type Registry struct {
	commands map[string]Command
}

// NewRegistry initializes an empty command registry.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register adds one or more commands to the registry.
func (r *Registry) Register(cmds ...Command) {
	for _, cmd := range cmds {
		name := strings.ToLower(cmd.Name())
		r.commands[name] = cmd

		for _, alias := range cmd.Aliases() {
			aliasName := strings.ToLower(alias)
			r.commands[aliasName] = cmd
		}
	}
}

// CommandsReturns a list of all uniquely registered commands.
func (r *Registry) Commands() []Command {
	unique := make(map[string]Command)
	for _, cmd := range r.commands {
		unique[strings.ToLower(cmd.Name())] = cmd
	}

	var list []Command
	for _, cmd := range unique {
		list = append(list, cmd)
	}
	return list
}

// Dispatch finds and runs a command, evaluating guards first.
//
// Key design decisions:
// - Centralized try/catch (error handling log) for command execution.
// - Guards are evaluated before Execute(), aborting early if failed.
// - Returns *CommandResult so middleware can inspect the status structuredly.
func (r *Registry) Dispatch(ctx *Context) (*CommandResult, error) {
	if !ctx.Msg.IsCommand {
		return nil, nil
	}

	cmd, exists := r.commands[ctx.Msg.CommandName]
	if !exists {
		// Command not found, silently ignore to not spam users.
		return &CommandResult{
			Status:  StatusNotFound,
			Command: ctx.Msg.CommandName,
		}, nil
	}

	// Run all guards attached to the command
	for _, g := range cmd.Guards() {
		allowed, reason := g.Check(ctx.Msg)
		if !allowed {
			if reason != "" {
				_ = ctx.Reply(reason)
			}
			ctx.Logger.Warn("Command rejected by guard",
				zap.String("command", cmd.Name()),
				zap.String("reason", reason),
				zap.String("sender", ctx.Msg.From.String()),
			)
			return &CommandResult{
				Status:  StatusGuardDenied,
				Command: cmd.Name(),
			}, nil
		}
	}

	// Call Execute()
	if err := cmd.Execute(ctx); err != nil {
		ctx.Logger.Error("Command failed",
			zap.String("command", cmd.Name()),
			zap.Error(err),
		)
		_ = ctx.Reply("⚠️ Something went wrong while executing the command.")
		return &CommandResult{
			Status:  StatusError,
			Command: cmd.Name(),
			Err:     err,
		}, err
	}

	return &CommandResult{
		Status:  StatusOk,
		Command: cmd.Name(),
	}, nil
}

// DispatchHandler adapts Dispatch to middleware.HandlerFunc and sets ctx.Result
// so that outer middleware (like Logger) can inspect the output status.
func (r *Registry) DispatchHandler() func(ctx *Context) error {
	return func(ctx *Context) error {
		res, err := r.Dispatch(ctx)
		ctx.Result = res
		return err
	}
}
