package guard

import (
	"luma/internal/parser"
)

// Guard is a pre-execution check attached per-command.
// Note: Changed the parameter from *command.Context to *parser.ParsedMessage
// to avoid a cyclic dependency between the guard and command packages.
type Guard interface {
	Check(msg *parser.ParsedMessage) (allowed bool, reason string)
}

// Key design decisions:
// - Guards act as reusable filters to protect commands.
// - Checking logic is decoupled from command execution logic.
// - Returns a user-friendly reason on failure to support graceful rejection.
