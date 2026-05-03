package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildStepEvent_StreamEvent(t *testing.T) {
	var _ StreamEvent = BuildStepEvent{}
}

func TestParseBuildLine(t *testing.T) {
	tests := []struct {
		input    string
		wantType string
		want     interface{}
	}{
		{
			input:    "#1 [internal] load .dockerignore",
			wantType: "BuildStepEvent",
			want: BuildStepEvent{
				Index:    1,
				Stage:    "internal",
				Step:     "load .dockerignore",
				Duration: "",
				Status:   "running",
			},
		},
		{
			input:    "#5 [stage 1/3] FROM alpine:latest",
			wantType: "BuildStepEvent",
			want: BuildStepEvent{
				Index:    5,
				Stage:    "stage 1/3",
				Step:     "FROM alpine:latest",
				Duration: "",
				Status:   "running",
			},
		},
		{
			input:    "#3 [build 2/5] COPY . /app    done",
			wantType: "BuildStepEvent",
			want: BuildStepEvent{
				Index:    3,
				Stage:    "build 2/5",
				Step:     "COPY . /app",
				Duration: "done",
				Status:   "done",
			},
		},
		{
			input:    "#7 [internal] load build context  0.1s",
			wantType: "BuildStepEvent",
			want: BuildStepEvent{
				Index:    7,
				Stage:    "internal",
				Step:     "load build context",
				Duration: "0.1s",
				Status:   "running",
			},
		},
		{
			input:    "#2 [stage 1/2] RUN apk add bash  CACHED",
			wantType: "BuildStepEvent",
			want: BuildStepEvent{
				Index:    2,
				Stage:    "stage 1/2",
				Step:     "RUN apk add bash",
				Duration: "CACHED",
				Status:   "cached",
			},
		},
		{
			input:    "Some random build output",
			wantType: "RawLine",
			want:     RawLine{Text: "Some random build output"},
		},
		{
			input:    "#10 exporting to image",
			wantType: "BuildStepEvent",
			want: BuildStepEvent{
				Index:    10,
				Stage:    "",
				Step:     "exporting to image",
				Duration: "",
				Status:   "running",
			},
		},
		{
			input:    "",
			wantType: "RawLine",
			want:     RawLine{Text: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			events := parseBuildLine(tt.input)
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}

			switch tt.wantType {
			case "BuildStepEvent":
				got, ok := events[0].(BuildStepEvent)
				if !ok {
					t.Fatalf("expected BuildStepEvent, got %T", events[0])
				}
				want := tt.want.(BuildStepEvent)
				if got != want {
					t.Errorf("parseBuildLine(%q) = %+v, want %+v", tt.input, got, want)
				}
			case "RawLine":
				got, ok := events[0].(RawLine)
				if !ok {
					t.Fatalf("expected RawLine, got %T", events[0])
				}
				want := tt.want.(RawLine)
				if got != want {
					t.Errorf("parseBuildLine(%q) = %+v, want %+v", tt.input, got, want)
				}
			}
		})
	}
}

func TestDefaultClient_StreamBuild(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "build.sh")

	script := `#!/bin/bash
if [[ "$1" == "build" ]]; then
  echo "#1 [internal] load build context"
  echo "#2 [stage 1/2] FROM alpine:latest"
  echo "#2 [stage 1/2] FROM alpine:latest  0.5s"
  echo "#3 [stage 2/2] RUN echo hello  done"
  echo "Successfully tagged test:latest"
  exit 0
fi
exit 1
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	client := NewDefaultClient(WithBinary(scriptPath))

	opts := BuildOpts{
		ContextPath: ".",
		Tag:         "test:latest",
	}

	stream, err := client.StreamBuild(ctx, opts)
	if err != nil {
		t.Fatalf("StreamBuild failed: %v", err)
	}

	events, result := drainStream(t, stream)
	if result.Err != nil {
		t.Errorf("unexpected error: %v", result.Err)
	}

	var steps []BuildStepEvent
	var raw []RawLine
	for _, ev := range events {
		switch e := ev.(type) {
		case BuildStepEvent:
			steps = append(steps, e)
		case RawLine:
			raw = append(raw, e)
		}
	}

	if len(steps) < 3 {
		t.Errorf("expected at least 3 build steps, got %d: %v", len(steps), steps)
	}

	if len(raw) < 1 {
		t.Errorf("expected at least 1 raw line, got %d", len(raw))
	}

	if len(steps) > 0 {
		if steps[0].Index != 1 || steps[0].Stage != "internal" {
			t.Errorf("unexpected first step: %+v", steps[0])
		}
	}
}

func TestBuildOpts_Args(t *testing.T) {
	tests := []struct {
		name string
		opts BuildOpts
		want []string
	}{
		{
			name: "minimal",
			opts: BuildOpts{
				ContextPath: ".",
			},
			want: []string{"build", "."},
		},
		{
			name: "with tag",
			opts: BuildOpts{
				ContextPath: "/path/to/build",
				Tag:         "myimage:v1",
			},
			want: []string{"build", "/path/to/build", "-t", "myimage:v1"},
		},
		{
			name: "with containerfile",
			opts: BuildOpts{
				ContextPath:       ".",
				ContainerfilePath: "Dockerfile.dev",
			},
			want: []string{"build", ".", "-f", "Dockerfile.dev"},
		},
		{
			name: "with platform",
			opts: BuildOpts{
				ContextPath: ".",
				Platform:    "linux/arm64",
			},
			want: []string{"build", ".", "--platform", "linux/arm64"},
		},
		{
			name: "with extra args",
			opts: BuildOpts{
				ContextPath: ".",
				ExtraArgs:   []string{"--no-cache", "--pull"},
			},
			want: []string{"build", ".", "--no-cache", "--pull"},
		},
		{
			name: "full options",
			opts: BuildOpts{
				ContextPath:       "/src",
				Tag:               "app:prod",
				ContainerfilePath: "Containerfile",
				Platform:          "linux/amd64",
				ExtraArgs:         []string{"--quiet"},
			},
			want: []string{"build", "/src", "-t", "app:prod", "-f", "Containerfile", "--platform", "linux/amd64", "--quiet"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.args()
			if len(got) != len(tt.want) {
				t.Errorf("args() len = %d, want %d\ngot:  %v\nwant: %v", len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("args()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFake_StreamBuild(t *testing.T) {
	fake := NewFake()

	events := []StreamEvent{
		BuildStepEvent{Index: 1, Stage: "internal", Step: "load", Status: "running"},
		BuildStepEvent{Index: 2, Stage: "stage 1/1", Step: "FROM alpine", Status: "done"},
	}
	fake.ReplayBuildStream(events, 0)

	opts := BuildOpts{ContextPath: "/test", Tag: "test:v1"}
	stream, err := fake.StreamBuild(context.Background(), opts)
	if err != nil {
		t.Fatalf("StreamBuild failed: %v", err)
	}

	// Check recorded call
	fake.mu.Lock()
	calls := fake.Calls
	fake.mu.Unlock()

	found := false
	for _, call := range calls {
		if call == "StreamBuild(context=/test,tag=test:v1)" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected call StreamBuild(context=/test,tag=test:v1), got: %v", calls)
	}

	// Verify events
	var steps []BuildStepEvent

	// Drain all events first
	for event := range stream.Events {
		if step, ok := event.(BuildStepEvent); ok {
			steps = append(steps, step)
		}
	}

	// Then check result
	result := <-stream.Done
	if result.Err != nil {
		t.Errorf("unexpected error: %v", result.Err)
	}

	if len(steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(steps))
	}
}
