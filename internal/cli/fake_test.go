package cli

import (
	"context"
	"errors"
	"testing"
)

func TestFakeCapabilitiesRecordsCall(t *testing.T) {
	f := &Fake{CapsResp: Capabilities{Version: "0.12.1", Major: 0, Minor: 12, Patch: 1}}
	got, err := f.Capabilities(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Version != "0.12.1" {
		t.Errorf("Version = %q, want 0.12.1", got.Version)
	}
	if len(f.Calls) != 1 || f.Calls[0] != "Capabilities" {
		t.Errorf("Calls = %v, want [Capabilities]", f.Calls)
	}
}

func TestFakeInspectContainer(t *testing.T) {
	f := &Fake{InspectContainerResp: []byte(`{"id":"x"}`)}
	got, err := f.InspectContainer(context.Background(), "x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(got)
	if s != `{"id":"x"}` {
		t.Errorf("got %q, want {\"id\":\"x\"}", s)
	}
	if len(f.Calls) != 1 || f.Calls[0] != "InspectContainer" {
		t.Errorf("Calls = %v, want [InspectContainer]", f.Calls)
	}
}

func TestFakeStopKillRestartDelete(t *testing.T) {
	f := &Fake{}
	ctx := context.Background()
	ops := []struct {
		name string
		call func() error
	}{
		{"StopContainer", func() error { return f.StopContainer(ctx, "x") }},
		{"KillContainer", func() error { return f.KillContainer(ctx, "x") }},
		{"RestartContainer", func() error { return f.RestartContainer(ctx, "x") }},
		{"DeleteContainer", func() error { return f.DeleteContainer(ctx, "x") }},
	}
	for _, op := range ops {
		if err := op.call(); err != nil {
			t.Errorf("%s: unexpected error: %v", op.name, err)
		}
	}
	if len(f.Calls) != 4 {
		t.Fatalf("Calls = %v, want length 4", f.Calls)
	}
	want := []string{"StopContainer", "KillContainer", "RestartContainer", "DeleteContainer"}
	for i, w := range want {
		if f.Calls[i] != w {
			t.Errorf("Calls[%d] = %q, want %q", i, f.Calls[i], w)
		}
	}
}

func TestFakePauseUnpause(t *testing.T) {
	f := &Fake{}
	if err := f.PauseContainer(context.Background(), "x"); err != nil {
		t.Errorf("PauseContainer should be a no-op when no err set: %v", err)
	}
	if err := f.UnpauseContainer(context.Background(), "x"); err != nil {
		t.Errorf("UnpauseContainer should be a no-op when no err set: %v", err)
	}
	if len(f.Calls) != 2 {
		t.Fatalf("Calls = %v, want length 2", f.Calls)
	}
	if f.Calls[0] != "PauseContainer" || f.Calls[1] != "UnpauseContainer" {
		t.Errorf("Calls = %v, want [PauseContainer, UnpauseContainer]", f.Calls)
	}
}

func TestFakePruneContainers(t *testing.T) {
	f := &Fake{PruneContainersResp: 5}
	n, err := f.PruneContainers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Errorf("n = %d, want 5", n)
	}
	if len(f.Calls) != 1 || f.Calls[0] != "PruneContainers" {
		t.Errorf("Calls = %v, want [PruneContainers]", f.Calls)
	}
}

func TestFakeListContainersErrOnly(t *testing.T) {
	sentinel := errors.New("boom")
	f := &Fake{ListContainersErr: sentinel}
	if _, err := f.ListContainers(context.Background(), false); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
