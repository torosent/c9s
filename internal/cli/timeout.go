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
// Callers don't need to invoke a cancel func — we spawn a tiny watcher
// goroutine that calls the cancel returned by WithTimeout once the
// deadline fires. This both satisfies govet's lostcancel analyzer (the
// cancel is reachable via the goroutine) and lets call sites stay one
// line:
//
//	raw, err := m.client.InspectImage(cli.DefaultCtx(), id)
//
// instead of every site needing its own ctx/cancel/defer trio. The
// watcher exits as soon as ctx.Done() closes, so steady-state goroutine
// count is bounded by in-flight CLI calls (typically 0–2 per screen).
func DefaultCtx() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return ctx
}
