package command

// DispatchStatus represents the outcome of a command dispatch.
type DispatchStatus string

const (
	StatusOk          DispatchStatus = "ok"
	StatusGuardDenied DispatchStatus = "guard_denied"
	StatusNotFound    DispatchStatus = "not_found"
	StatusError       DispatchStatus = "error"
)

// CommandResult carries structured information about command execution.
// It is returned by Dispatch and injected into Context for middleware to read.
type CommandResult struct {
	Status  DispatchStatus
	Command string
	Err     error
}

// IsError returns true if the result represents an error.
func (r *CommandResult) IsError() bool {
	return r.Status == StatusError
}
