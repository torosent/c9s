package state

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
)

// RefreshedMsg is emitted when a resource fetch completes.
type RefreshedMsg[T any] struct {
	Resource cli.Resource
	Snapshot Snapshot[T]
}

// TickMsg is emitted when a clock tick fires. Screens should switch on this
// to schedule periodic refreshes.
type TickMsg struct {
	Resource cli.Resource
	T        time.Time
}

// TickCmd returns a tea.Cmd that blocks until the clock ticks once,
// then emits a TickMsg.
func TickCmd(d time.Duration, clk clock.Clock, resource cli.Resource) tea.Cmd {
	return func() tea.Msg {
		t := <-clk.Tick(d)
		return TickMsg{Resource: resource, T: t}
	}
}

// MakeRefreshedCmd returns a tea.Cmd that fetches items and emits
// a RefreshedMsg with the results.
func MakeRefreshedCmd[T any](ctx context.Context, fetch func(context.Context) ([]T, error), resource cli.Resource) tea.Cmd {
	return func() tea.Msg {
		items, err := fetch(ctx)
		return RefreshedMsg[T]{
			Resource: resource,
			Snapshot: Snapshot[T]{
				Items:     items,
				FetchedAt: time.Now(),
				Err:       err,
			},
		}
	}
}
