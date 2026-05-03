package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRawLine_StreamEvent(t *testing.T) {
	var _ StreamEvent = RawLine{}
}

func TestStream_BasicFlow(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.sh")

	script := `#!/bin/bash
echo "line 1"
echo "line 2"
echo "line 3"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	cmd := exec.CommandContext(ctx, "/bin/bash", scriptPath)

	parse := func(line string) []StreamEvent {
		return []StreamEvent{RawLine{Text: line}}
	}

	stream, err := runStream(ctx, cmd, parse)
	if err != nil {
		t.Fatalf("runStream failed: %v", err)
	}

	events, result := drainStream(t, stream)
	if result.Err != nil {
		t.Errorf("unexpected error: %v", result.Err)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	var lines []string
	for _, ev := range events {
		if raw, ok := ev.(RawLine); ok {
			lines = append(lines, raw.Text)
		}
	}

	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(lines), lines)
	}

	expected := []string{"line 1", "line 2", "line 3"}
	for i, want := range expected {
		if i >= len(lines) || lines[i] != want {
			t.Errorf("line %d: expected %q, got %q", i, want, lines[i])
		}
	}
}

func TestStream_NonZeroExit(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "fail.sh")

	script := `#!/bin/bash
echo "before error"
exit 42
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	cmd := exec.CommandContext(ctx, "/bin/bash", scriptPath)

	parse := func(line string) []StreamEvent {
		return []StreamEvent{RawLine{Text: line}}
	}

	stream, err := runStream(ctx, cmd, parse)
	if err != nil {
		t.Fatalf("runStream failed: %v", err)
	}

	events, result := drainStream(t, stream)
	if result.Err == nil {
		t.Error("expected error for non-zero exit, got nil")
	}
	if result.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", result.ExitCode)
	}

	gotLine := false
	for _, ev := range events {
		if raw, ok := ev.(RawLine); ok && raw.Text == "before error" {
			gotLine = true
		}
	}

	if !gotLine {
		t.Error("expected to see 'before error' line")
	}
}

func TestStream_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "sleep.sh")

	script := `#!/bin/bash
trap 'echo "got sigint"; exit 130' INT
echo "starting"
sleep 10
echo "finished"
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	cmd := exec.CommandContext(ctx, "/bin/bash", scriptPath)

	parse := func(line string) []StreamEvent {
		return []StreamEvent{RawLine{Text: line}}
	}

	stream, err := runStream(ctx, cmd, parse)
	if err != nil {
		t.Fatalf("runStream failed: %v", err)
	}

	var sawStarting bool
	var sawSigint bool

	// Wait for "starting" line
	for event := range stream.Events {
		if raw, ok := event.(RawLine); ok && strings.Contains(raw.Text, "starting") {
			sawStarting = true
			break
		}
	}

	if !sawStarting {
		t.Fatal("never saw 'starting' line")
	}

	// Cancel context
	cancel()

	events, result := drainStream(t, stream)
	if result.Err == nil {
		t.Error("expected error after cancellation")
	}

	for _, ev := range events {
		if raw, ok := ev.(RawLine); ok && strings.Contains(raw.Text, "got sigint") {
			sawSigint = true
		}
	}

	if !sawSigint {
		t.Log("note: script may not have received SIGINT in time")
	}
}

func TestStream_ParserReturnsMultipleEvents(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "multi.sh")

	script := `#!/bin/bash
echo "single"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	cmd := exec.CommandContext(ctx, "/bin/bash", scriptPath)

	// Parser that returns multiple events per line
	parse := func(line string) []StreamEvent {
		return []StreamEvent{
			RawLine{Text: line + "-1"},
			RawLine{Text: line + "-2"},
		}
	}

	stream, err := runStream(ctx, cmd, parse)
	if err != nil {
		t.Fatalf("runStream failed: %v", err)
	}

	events, _ := drainStream(t, stream)
	var lines []string
	for _, ev := range events {
		if raw, ok := ev.(RawLine); ok {
			lines = append(lines, raw.Text)
		}
	}

	if len(lines) != 2 {
		t.Errorf("expected 2 events, got %d: %v", len(lines), lines)
	}
	if len(lines) >= 2 {
		if lines[0] != "single-1" || lines[1] != "single-2" {
			t.Errorf("unexpected events: %v", lines)
		}
	}
}

func TestStream_ChannelsClosedOnCompletion(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "quick.sh")

	script := `#!/bin/bash
echo "done"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	cmd := exec.CommandContext(ctx, "/bin/bash", scriptPath)

	parse := func(line string) []StreamEvent {
		return []StreamEvent{RawLine{Text: line}}
	}

	stream, err := runStream(ctx, cmd, parse)
	if err != nil {
		t.Fatalf("runStream failed: %v", err)
	}

	// Drain events
	for range stream.Events {
	}

	// Events channel should be closed
	_, ok := <-stream.Events
	if ok {
		t.Error("Events channel not closed after completion")
	}

	// Done should have a result
	result, ok := <-stream.Done
	if !ok {
		t.Error("Done channel closed without result")
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d", result.ExitCode)
	}

	// Done channel should also be closed
	_, ok = <-stream.Done
	if ok {
		t.Error("Done channel not closed")
	}
}
