package cli

import (
	"context"
	"time"
)

// DefaultRequestTimeout bounds individual `container` CLI calls. It needs
// to be long enough for the runtime to respond when it's busy (e.g. while
// other writes are in flight) but short enough that a stuck process gets
// torn down before the next refresh tick fires another duplicate call.
const DefaultRequestTimeout = 5 * time.Second

// DefaultCtx returns a context with a DefaultRequestTimeout deadline.
// Callers don't need to call the returned cancel func: context.WithTimeout
// schedules an internal timer that automatically cancels the context and
// releases resources when the deadline fires. Dropping the cancel keeps
// call sites compact (no `defer cancel()` boilerplate at every screen
// closure), at the cost of allowing the timer to run to expiration even
// if the call returns early. That's acceptable for our 5 s bound.
//
//nolint:govet // intentional: cancel is auto-fired by the deadline timer.
func DefaultCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	return ctx
}
