package discord

import "fmt"

// CommandError represents an error handling a Discord command.
type CommandError struct {
	UserMsg string // user-facing error message. will be sent back as feedback via Discord interaction response
	Err     error  // wrapped error. internal
}

// CErr returns a CommandError.
func CErr(userMsg string, err error) *CommandError {
	return &CommandError{UserMsg: userMsg, Err: err}
}

// Error returns error string for CommandError.
func (e *CommandError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.UserMsg, e.Err)
	}
	return e.UserMsg
}

// Unwrap unwraps a CommandError.
func (e *CommandError) Unwrap() error { return e.Err }
