package cli

import (
	"errors"
	"fmt"
)

// ErrUnsupported is returned when the underlying `container` CLI does not
// implement a requested operation (e.g., on older versions). Callers can
// detect it with errors.Is(err, cli.ErrUnsupported).
var ErrUnsupported = errors.New("operation not supported by this container CLI version")

// Error wraps an underlying CLI failure with structured context.
type Error struct {
	Op       string // e.g. "cli.list-containers"
	Resource string // e.g. "container/api-server"
	Cause    error
	Hint     string // optional human-readable suggestion
}

// Error implements error.
func (e *Error) Error() string {
	base := fmt.Sprintf("%s on %s: %v", e.Op, e.Resource, e.Cause)
	if e.Hint != "" {
		return base + " (hint: " + e.Hint + ")"
	}
	return base
}

// Unwrap exposes the underlying cause for errors.Is / errors.As.
func (e *Error) Unwrap() error { return e.Cause }

// Wrap returns a new *Error, or nil if cause is nil.
func Wrap(op, resource string, cause error, hint string) error {
	if cause == nil {
		return nil
	}
	return &Error{Op: op, Resource: resource, Cause: cause, Hint: hint}
}
