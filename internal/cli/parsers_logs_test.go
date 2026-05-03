package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLogLine_StreamEvent(t *testing.T) {
	var _ StreamEvent = LogLine{}
}

func TestParseLogLine_WithLevels(t *testing.T) {
	tests := []struct {
		input string
		want  LogLine
	}{
		{"INFO: server started", LogLine{Raw: "INFO: server started", Level: "INFO"}},
		{"WARN connection timeout", LogLine{Raw: "WARN connection timeout", Level: "WARN"}},
		{"ERROR: failed to connect", LogLine{Raw: "ERROR: failed to connect", Level: "ERROR"}},
		{"DEBUG trace point", LogLine{Raw: "DEBUG trace point", Level: "DEBUG"}},
		{"info lowercase", LogLine{Raw: "info lowercase", Level: "INFO"}},
		{"Error mixed case", LogLine{Raw: "Error mixed case", Level: "ERROR"}},
		{"plain log line", LogLine{Raw: "plain log line", Level: ""}},
		{"NOTLEVEL: something", LogLine{Raw: "NOTLEVEL: something", Level: ""}},
		{"", LogLine{Raw: "", Level: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			events := parseLogLine(tt.input)
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			got, ok := events[0].(LogLine)
			if !ok {
				t.Fatalf("expected LogLine, got %T", events[0])
			}
			if got != tt.want {
				t.Errorf("parseLogLine(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestDefaultClient_StreamLogs(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "logs.sh")

	script := `#!/bin/bash
# Args: $1=logs $2=test-id $3=--follow (optional)
if [[ "$1" == "logs" ]]; then
  echo "INFO: starting container"
  echo "WARN: low memory"
  echo "plain log"
  if [[ "$3" == "--follow" ]]; then
    echo "ERROR: connection lost"
  fi
  exit 0
fi
exit 1
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	client := NewDefaultClient(WithBinary(scriptPath))

	t.Run("follow=false", func(t *testing.T) {
		stream, err := client.StreamLogs(ctx, "test-id", false)
		if err != nil {
			t.Fatalf("StreamLogs failed: %v", err)
		}

		events, result := drainStream(t, stream)
		if result.Err != nil {
			t.Errorf("unexpected error: %v", result.Err)
		}

		var logs []LogLine
		for _, ev := range events {
			if log, ok := ev.(LogLine); ok {
				logs = append(logs, log)
			}
		}

		if len(logs) != 3 {
			t.Errorf("expected 3 logs, got %d: %v", len(logs), logs)
		}

		want := []LogLine{
			{Raw: "INFO: starting container", Level: "INFO"},
			{Raw: "WARN: low memory", Level: "WARN"},
			{Raw: "plain log", Level: ""},
		}

		for i, w := range want {
			if i >= len(logs) || logs[i] != w {
				t.Errorf("log %d: got %+v, want %+v", i, logs[i], w)
			}
		}
	})

	t.Run("follow=true", func(t *testing.T) {
		stream, err := client.StreamLogs(ctx, "test-id", true)
		if err != nil {
			t.Fatalf("StreamLogs failed: %v", err)
		}

		events, result := drainStream(t, stream)
		if result.Err != nil {
			t.Errorf("unexpected error: %v", result.Err)
		}

		var logs []LogLine
		for _, ev := range events {
			if log, ok := ev.(LogLine); ok {
				logs = append(logs, log)
			}
		}

		if len(logs) != 4 {
			t.Errorf("expected 4 logs, got %d: %v", len(logs), logs)
		}

		// Should have the ERROR line when --follow is passed
		hasError := false
		for _, log := range logs {
			if log.Level == "ERROR" {
				hasError = true
				break
			}
		}
		if !hasError {
			t.Error("expected ERROR log when follow=true")
		}
	})
}

func TestFake_StreamLogs(t *testing.T) {
	fake := NewFake()

	events := []StreamEvent{
		LogLine{Raw: "INFO: test", Level: "INFO"},
		LogLine{Raw: "plain", Level: ""},
	}
	fake.ReplayLogStream(events, 0)

	stream, err := fake.StreamLogs(context.Background(), "test-id", true)
	if err != nil {
		t.Fatalf("StreamLogs failed: %v", err)
	}

	// Check recorded call
	fake.mu.Lock()
	calls := fake.Calls
	fake.mu.Unlock()

	found := false
	for _, call := range calls {
		if call == "StreamLogs(test-id,follow=true)" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected call StreamLogs(test-id,follow=true), got: %v", calls)
	}

	// Drain events first; the channel is closed by the producer goroutine
	// before Done is signalled. Reading from Done in the same select races
	// with the buffered Events channel.
	var logs []LogLine
	for event := range stream.Events {
		if log, ok := event.(LogLine); ok {
			logs = append(logs, log)
		}
	}

	result := <-stream.Done
	if result.Err != nil {
		t.Errorf("unexpected error: %v", result.Err)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	if len(logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(logs))
	}
}
