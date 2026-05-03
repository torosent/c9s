package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRunOptsArgs(t *testing.T) {
	opts := RunOpts{
		Name:        "api",
		Image:       "ghcr.io/acme/api:1.0",
		Ports:       []string{"8080:8080", ""},
		Env:         []string{"KEY=val"},
		Volumes:     []string{"data:/data"},
		Interactive: true,
		TTY:         true,
		Detach:      true,
	}
	args := opts.args()
	expected := []string{
		"run", "--name", "api",
		"-p", "8080:8080",
		"-e", "KEY=val",
		"-v", "data:/data",
		"-i", "-t", "-d",
		"ghcr.io/acme/api:1.0",
	}
	if len(args) != len(expected) {
		t.Fatalf("len(args) = %d, want %d (%v)", len(args), len(expected), args)
	}
	for i, want := range expected {
		if args[i] != want {
			t.Errorf("args[%d] = %q, want %q (full: %v)", i, args[i], want, args)
		}
	}
}

func TestRunOptsArgsMinimal(t *testing.T) {
	opts := RunOpts{Image: "alpine"}
	args := opts.args()
	if len(args) != 2 || args[0] != "run" || args[1] != "alpine" {
		t.Errorf("got %v", args)
	}
}

func TestRunContainerStreams(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	script := "#!/bin/bash\necho 'started'\nexit 0\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	stream, err := c.RunContainer(context.Background(), RunOpts{Image: "alpine"})
	if err != nil {
		t.Fatalf("RunContainer: %v", err)
	}
	count := 0
	for range stream.Events {
		count++
	}
	<-stream.Done
	if count == 0 {
		t.Errorf("expected at least one event from RunContainer")
	}
}

func TestParseRunLineEmpty(t *testing.T) {
	ev := parseRunLine("")
	if len(ev) != 1 {
		t.Errorf("expected 1 event, got %d", len(ev))
	}
}

func TestFakeRunContainer(t *testing.T) {
	f := NewFake()
	f.ReplayRunStream([]StreamEvent{RawLine{Text: "ok"}}, 0)
	stream, err := f.RunContainer(context.Background(), RunOpts{Image: "alpine"})
	if err != nil {
		t.Fatalf("RunContainer: %v", err)
	}
	count := 0
	for range stream.Events {
		count++
	}
	res := <-stream.Done
	if res.ExitCode != 0 {
		t.Errorf("ExitCode = %d", res.ExitCode)
	}
	if count != 1 {
		t.Errorf("expected 1 event")
	}
}
