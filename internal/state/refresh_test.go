package state

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
)

func TestTickCmdFires(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC))
	cmd := TickCmd(2*time.Second, clk, cli.ResourceContainers)

	// Invoke cmd in a goroutine
	resultCh := make(chan tea.Msg, 1)
	go func() {
		resultCh <- cmd()
	}()

	// Command should block initially
	select {
	case <-resultCh:
		t.Fatal("cmd returned before clock advanced")
	case <-time.After(100 * time.Millisecond):
		// Expected: cmd is blocked
	}

	// Advance clock
	clk.Advance(2 * time.Second)

	// Now we should get the message
	select {
	case msg := <-resultCh:
		tick, ok := msg.(TickMsg)
		if !ok {
			t.Fatalf("expected TickMsg, got %T", msg)
		}
		if tick.Resource != cli.ResourceContainers {
			t.Errorf("expected resource 'containers', got '%s'", tick.Resource)
		}
		if tick.T.IsZero() {
			t.Error("expected non-zero time")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("cmd did not return after clock advance")
	}
}

func TestMakeRefreshedCmdSuccess(t *testing.T) {
	fetch := func(ctx context.Context) ([]int, error) {
		return []int{1, 2, 3}, nil
	}

	cmd := MakeRefreshedCmd[int](context.Background(), fetch, cli.ResourceContainers)
	msg := cmd()

	refreshed, ok := msg.(RefreshedMsg[int])
	if !ok {
		t.Fatalf("expected RefreshedMsg[int], got %T", msg)
	}

	if refreshed.Resource != cli.ResourceContainers {
		t.Errorf("expected resource 'containers', got '%s'", refreshed.Resource)
	}

	if len(refreshed.Snapshot.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(refreshed.Snapshot.Items))
	}

	if refreshed.Snapshot.Items[0] != 1 || refreshed.Snapshot.Items[1] != 2 || refreshed.Snapshot.Items[2] != 3 {
		t.Errorf("unexpected items: %v", refreshed.Snapshot.Items)
	}

	if refreshed.Snapshot.Err != nil {
		t.Errorf("expected no error, got %v", refreshed.Snapshot.Err)
	}

	if refreshed.Snapshot.FetchedAt.IsZero() {
		t.Error("expected non-zero FetchedAt")
	}
}

func TestMakeRefreshedCmdError(t *testing.T) {
	sentinel := errors.New("fetch failed")
	fetch := func(ctx context.Context) ([]int, error) {
		return nil, sentinel
	}

	cmd := MakeRefreshedCmd[int](context.Background(), fetch, cli.ResourceContainers)
	msg := cmd()

	refreshed, ok := msg.(RefreshedMsg[int])
	if !ok {
		t.Fatalf("expected RefreshedMsg[int], got %T", msg)
	}

	if refreshed.Resource != cli.ResourceContainers {
		t.Errorf("expected resource 'containers', got '%s'", refreshed.Resource)
	}

	if !errors.Is(refreshed.Snapshot.Err, sentinel) {
		t.Errorf("expected sentinel error, got %v", refreshed.Snapshot.Err)
	}

	if len(refreshed.Snapshot.Items) != 0 {
		t.Errorf("expected empty items on error, got %d", len(refreshed.Snapshot.Items))
	}
}
