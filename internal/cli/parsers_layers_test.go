package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLayerProgress_StreamEvent(t *testing.T) {
	var _ StreamEvent = LayerProgressEvent{}
}

func TestParseLayerLine_Pull(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
		check    func(t *testing.T, events []StreamEvent)
	}{
		{
			name:     "downloading line",
			input:    "sha256:abc123 downloading 1024000 / 10240000",
			wantType: "LayerProgressEvent",
			check: func(t *testing.T, events []StreamEvent) {
				if len(events) != 1 {
					t.Fatalf("expected 1 event, got %d", len(events))
				}
				e, ok := events[0].(LayerProgressEvent)
				if !ok {
					t.Fatalf("expected LayerProgressEvent, got %T", events[0])
				}
				if len(e.Layers) != 1 {
					t.Fatalf("expected 1 layer, got %d", len(e.Layers))
				}
				layer := e.Layers[0]
				if layer.Digest != "sha256:abc123" {
					t.Errorf("digest = %q, want sha256:abc123", layer.Digest)
				}
				if layer.State != "downloading" {
					t.Errorf("state = %q, want downloading", layer.State)
				}
				if layer.BytesDone != 1024000 {
					t.Errorf("bytesDone = %d, want 1024000", layer.BytesDone)
				}
				if layer.BytesTotal != 10240000 {
					t.Errorf("bytesTotal = %d, want 10240000", layer.BytesTotal)
				}
			},
		},
		{
			name:     "extracting line",
			input:    "sha256:def456 extracting",
			wantType: "LayerProgressEvent",
			check: func(t *testing.T, events []StreamEvent) {
				if len(events) != 1 {
					t.Fatalf("expected 1 event, got %d", len(events))
				}
				e, ok := events[0].(LayerProgressEvent)
				if !ok {
					t.Fatalf("expected LayerProgressEvent, got %T", events[0])
				}
				if len(e.Layers) != 1 {
					t.Fatalf("expected 1 layer, got %d", len(e.Layers))
				}
				layer := e.Layers[0]
				if layer.State != "extracting" {
					t.Errorf("state = %q, want extracting", layer.State)
				}
			},
		},
		{
			name:     "done line",
			input:    "sha256:ghi789 done",
			wantType: "LayerProgressEvent",
			check: func(t *testing.T, events []StreamEvent) {
				if len(events) != 1 {
					t.Fatalf("expected 1 event, got %d", len(events))
				}
				e, ok := events[0].(LayerProgressEvent)
				if !ok {
					t.Fatalf("expected LayerProgressEvent, got %T", events[0])
				}
				layer := e.Layers[0]
				if layer.State != "done" {
					t.Errorf("state = %q, want done", layer.State)
				}
			},
		},
		{
			name:     "unparseable line",
			input:    "Pulling from registry...",
			wantType: "RawLine",
			check: func(t *testing.T, events []StreamEvent) {
				if len(events) != 1 {
					t.Fatalf("expected 1 event, got %d", len(events))
				}
				raw, ok := events[0].(RawLine)
				if !ok {
					t.Fatalf("expected RawLine, got %T", events[0])
				}
				if raw.Text != "Pulling from registry..." {
					t.Errorf("text = %q, want 'Pulling from registry...'", raw.Text)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := newLayerParser()
			events := parser.parseLine(tt.input)
			tt.check(t, events)
		})
	}
}

func TestParseLayerLine_StateAccumulation(t *testing.T) {
	parser := newLayerParser()

	// Simulate a sequence of updates for multiple layers
	inputs := []string{
		"sha256:aaa downloading 100 / 1000",
		"sha256:bbb downloading 200 / 2000",
		"sha256:aaa downloading 500 / 1000",
		"sha256:aaa extracting",
		"sha256:bbb done",
		"sha256:aaa done",
	}

	var lastEvent LayerProgressEvent
	for _, input := range inputs {
		events := parser.parseLine(input)
		if len(events) == 1 {
			if e, ok := events[0].(LayerProgressEvent); ok {
				lastEvent = e
			}
		}
	}

	// After all updates, we should have 2 layers
	if len(lastEvent.Layers) != 2 {
		t.Fatalf("expected 2 layers, got %d", len(lastEvent.Layers))
	}

	// Find layer aaa
	var layerA, layerB *LayerProgress
	for i := range lastEvent.Layers {
		if lastEvent.Layers[i].Digest == "sha256:aaa" {
			layerA = &lastEvent.Layers[i]
		}
		if lastEvent.Layers[i].Digest == "sha256:bbb" {
			layerB = &lastEvent.Layers[i]
		}
	}

	if layerA == nil || layerB == nil {
		t.Fatalf("missing layers: a=%v, b=%v", layerA, layerB)
	}

	if layerA.State != "done" {
		t.Errorf("layerA.State = %q, want done", layerA.State)
	}
	if layerA.BytesDone != 500 {
		t.Errorf("layerA.BytesDone = %d, want 500", layerA.BytesDone)
	}
	if layerB.State != "done" {
		t.Errorf("layerB.State = %q, want done", layerB.State)
	}
}

func TestParseLayerLine_Push(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, events []StreamEvent)
	}{
		{
			name:  "mounted layer",
			input: "sha256:xyz mounted",
			check: func(t *testing.T, events []StreamEvent) {
				e := events[0].(LayerProgressEvent)
				layer := e.Layers[0]
				if layer.State != "mounted" {
					t.Errorf("state = %q, want mounted", layer.State)
				}
				if !layer.Mounted {
					t.Error("expected Mounted = true")
				}
			},
		},
		{
			name:  "pushing layer",
			input: "sha256:xyz pushing 5000 / 10000",
			check: func(t *testing.T, events []StreamEvent) {
				e := events[0].(LayerProgressEvent)
				layer := e.Layers[0]
				if layer.State != "pushing" {
					t.Errorf("state = %q, want pushing", layer.State)
				}
				if layer.BytesDone != 5000 {
					t.Errorf("bytesDone = %d, want 5000", layer.BytesDone)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := newLayerParser()
			events := parser.parseLine(tt.input)
			tt.check(t, events)
		})
	}
}

func TestDefaultClient_StreamPull(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "pull.sh")

	script := `#!/bin/bash
if [[ "$1" == "image" && "$2" == "pull" ]]; then
  echo "Pulling from registry..."
  echo "sha256:layer1 downloading 100 / 1000"
  echo "sha256:layer1 extracting"
  echo "sha256:layer1 done"
  echo "Pull complete"
  exit 0
fi
exit 1
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	client := NewDefaultClient(WithBinary(scriptPath))

	stream, err := client.StreamPull(ctx, "test:latest")
	if err != nil {
		t.Fatalf("StreamPull failed: %v", err)
	}

	var layerEvents []LayerProgressEvent
	var rawLines []RawLine

	for event := range stream.Events {
		switch e := event.(type) {
		case LayerProgressEvent:
			layerEvents = append(layerEvents, e)
		case RawLine:
			rawLines = append(rawLines, e)
		}
	}

	result := <-stream.Done
	if result.Err != nil {
		t.Errorf("unexpected error: %v", result.Err)
	}

	if len(layerEvents) < 3 {
		t.Errorf("expected at least 3 layer events, got %d", len(layerEvents))
	}

	if len(rawLines) < 2 {
		t.Errorf("expected at least 2 raw lines, got %d", len(rawLines))
	}
}

func TestDefaultClient_StreamPush(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "push.sh")

	script := `#!/bin/bash
if [[ "$1" == "image" && "$2" == "push" ]]; then
  echo "Pushing to registry..."
  echo "sha256:layer1 pushing 100 / 1000"
  echo "sha256:layer2 mounted"
  echo "sha256:layer1 done"
  echo "Push complete"
  exit 0
fi
exit 1
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	client := NewDefaultClient(WithBinary(scriptPath))

	stream, err := client.StreamPush(ctx, "test:latest")
	if err != nil {
		t.Fatalf("StreamPush failed: %v", err)
	}

	var layerEvents []LayerProgressEvent

	for event := range stream.Events {
		if e, ok := event.(LayerProgressEvent); ok {
			layerEvents = append(layerEvents, e)
		}
	}

	result := <-stream.Done
	if result.Err != nil {
		t.Errorf("unexpected error: %v", result.Err)
	}

	if len(layerEvents) < 3 {
		t.Errorf("expected at least 3 layer events, got %d", len(layerEvents))
	}

	// Check that we tracked a mounted layer
	foundMounted := false
	for _, e := range layerEvents {
		for _, layer := range e.Layers {
			if layer.Mounted {
				foundMounted = true
				break
			}
		}
	}
	if !foundMounted {
		t.Error("expected to find at least one mounted layer")
	}
}

func TestFake_StreamPull(t *testing.T) {
	fake := NewFake()

	events := []StreamEvent{
		LayerProgressEvent{Layers: []LayerProgress{
			{Digest: "sha256:aaa", State: "downloading", BytesDone: 100, BytesTotal: 1000},
		}},
		LayerProgressEvent{Layers: []LayerProgress{
			{Digest: "sha256:aaa", State: "done", BytesDone: 1000, BytesTotal: 1000},
		}},
	}
	fake.ReplayLayerStream(events, 0)

	stream, err := fake.StreamPull(context.Background(), "test:v1")
	if err != nil {
		t.Fatalf("StreamPull failed: %v", err)
	}

	// Check recorded call
	fake.mu.Lock()
	calls := fake.Calls
	fake.mu.Unlock()

	found := false
	for _, call := range calls {
		if call == "StreamPull(test:v1)" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected call StreamPull(test:v1), got: %v", calls)
	}

	// Verify events
	var layerEvents []LayerProgressEvent

	for event := range stream.Events {
		if e, ok := event.(LayerProgressEvent); ok {
			layerEvents = append(layerEvents, e)
		}
	}

	result := <-stream.Done
	if result.Err != nil {
		t.Errorf("unexpected error: %v", result.Err)
	}

	if len(layerEvents) != 2 {
		t.Errorf("expected 2 layer events, got %d", len(layerEvents))
	}
}

func TestFake_StreamPush(t *testing.T) {
	fake := NewFake()

	events := []StreamEvent{
		LayerProgressEvent{Layers: []LayerProgress{
			{Digest: "sha256:bbb", State: "pushing", BytesDone: 50, BytesTotal: 100},
		}},
	}
	fake.ReplayLayerStream(events, 0)

	stream, err := fake.StreamPush(context.Background(), "test:v2")
	if err != nil {
		t.Fatalf("StreamPush failed: %v", err)
	}

	// Check recorded call
	fake.mu.Lock()
	calls := fake.Calls
	fake.mu.Unlock()

	found := false
	for _, call := range calls {
		if call == "StreamPush(test:v2)" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected call StreamPush(test:v2), got: %v", calls)
	}

	// Drain events
	for range stream.Events {
	}

	<-stream.Done
}
