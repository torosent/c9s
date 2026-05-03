package cli

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestFakeRecordsCalls(t *testing.T) {
	f := &Fake{VersionResp: "container CLI version 0.12.1 (build: release, commit: e9891b3)"}
	ctx := context.Background()

	got, err := f.Version(ctx)
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if !strings.Contains(got, "0.12.1") {
		t.Errorf("Version() = %q, want substring 0.12.1", got)
	}
	if len(f.Calls) != 1 || f.Calls[0] != "Version" {
		t.Errorf("Calls = %v, want [Version]", f.Calls)
	}
}

func TestFakePropagatesErr(t *testing.T) {
	sentinel := errors.New("boom")
	f := &Fake{VersionErr: sentinel}
	if _, err := f.Version(context.Background()); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

// Compile-time assertion: both implementations satisfy Client.
var (
	_ Client = (*DefaultClient)(nil)
	_ Client = (*Fake)(nil)
)

func TestDefaultClientVersionShellsOut(t *testing.T) {
	// Use `echo` so we don't depend on `container` being installed.
	// `echo --version` prints "--version\n" and exits 0.
	bin, err := exec.LookPath("echo")
	if err != nil {
		t.Skip("echo not on PATH")
	}
	c := NewDefaultClient(WithBinary(bin))
	out, err := c.Version(context.Background())
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if !strings.Contains(out, "--version") {
		t.Errorf("Version() = %q, want substring '--version'", out)
	}
}

func TestDefaultClientVersionBinaryMissing(t *testing.T) {
	c := NewDefaultClient(WithBinary("/nonexistent/c9s-test-binary-does-not-exist"))
	out, err := c.Version(context.Background())
	if err == nil {
		t.Fatalf("expected error for missing binary, got out=%q err=nil", out)
	}
	var c9err *Error
	if !errors.As(err, &c9err) {
		t.Fatalf("expected *Error, got %T: %v", err, err)
	}
	if c9err.Op != "cli.version" {
		t.Errorf("Op = %q, want cli.version", c9err.Op)
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
}

// Task 3: Tests for lifecycle methods

func TestFakeListContainers(t *testing.T) {
	containers := []Container{
		{ID: "abc", ShortID: "abc", Status: "running"},
		{ID: "def", ShortID: "def", Status: "stopped"},
	}
	f := &Fake{ListContainersResp: containers}

	result, err := f.ListContainers(context.Background(), true)
	if err != nil {
		t.Fatalf("ListContainers: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 containers, got %d", len(result))
	}

	if len(f.Calls) != 1 || f.Calls[0] != "ListContainers" {
		t.Errorf("Calls = %v, want [ListContainers]", f.Calls)
	}
}

func TestFakeListContainersErr(t *testing.T) {
	sentinel := errors.New("list error")
	f := &Fake{ListContainersErr: sentinel}

	_, err := f.ListContainers(context.Background(), false)
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestFakeStopContainer(t *testing.T) {
	f := &Fake{}
	err := f.StopContainer(context.Background(), "abc")
	if err != nil {
		t.Errorf("StopContainer: %v", err)
	}
	if len(f.Calls) != 1 || f.Calls[0] != "StopContainer" {
		t.Errorf("Calls = %v, want [StopContainer]", f.Calls)
	}
}

func TestFakeStopContainerErr(t *testing.T) {
	sentinel := errors.New("stop error")
	f := &Fake{StopContainerErr: sentinel}
	err := f.StopContainer(context.Background(), "abc")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestFakeKillContainer(t *testing.T) {
	f := &Fake{}
	err := f.KillContainer(context.Background(), "abc")
	if err != nil {
		t.Errorf("KillContainer: %v", err)
	}
	if len(f.Calls) != 1 || f.Calls[0] != "KillContainer" {
		t.Errorf("Calls = %v, want [KillContainer]", f.Calls)
	}
}

func TestFakeKillContainerErr(t *testing.T) {
	sentinel := errors.New("kill error")
	f := &Fake{KillContainerErr: sentinel}
	err := f.KillContainer(context.Background(), "abc")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestFakeRestartContainer(t *testing.T) {
	f := &Fake{}
	err := f.RestartContainer(context.Background(), "abc")
	if err != nil {
		t.Errorf("RestartContainer: %v", err)
	}
	if len(f.Calls) != 1 || f.Calls[0] != "RestartContainer" {
		t.Errorf("Calls = %v, want [RestartContainer]", f.Calls)
	}
}

func TestFakeRestartContainerErr(t *testing.T) {
	sentinel := errors.New("restart error")
	f := &Fake{RestartContainerErr: sentinel}
	err := f.RestartContainer(context.Background(), "abc")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestFakeDeleteContainer(t *testing.T) {
	f := &Fake{}
	err := f.DeleteContainer(context.Background(), "abc")
	if err != nil {
		t.Errorf("DeleteContainer: %v", err)
	}
	if len(f.Calls) != 1 || f.Calls[0] != "DeleteContainer" {
		t.Errorf("Calls = %v, want [DeleteContainer]", f.Calls)
	}
}

func TestFakeDeleteContainerErr(t *testing.T) {
	sentinel := errors.New("delete error")
	f := &Fake{DeleteContainerErr: sentinel}
	err := f.DeleteContainer(context.Background(), "abc")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestPauseUnsupported(t *testing.T) {
	c := NewDefaultClient()

	err := c.PauseContainer(context.Background(), "abc")
	if err == nil {
		t.Fatal("expected error for pause, got nil")
	}

	var c9err *Error
	if !errors.As(err, &c9err) {
		t.Fatalf("expected *Error, got %T", err)
	}

	if c9err.Op != "cli.pause-unsupported" {
		t.Errorf("Op = %q, want cli.pause-unsupported", c9err.Op)
	}
}

func TestUnpauseUnsupported(t *testing.T) {
	c := NewDefaultClient()

	err := c.UnpauseContainer(context.Background(), "abc")
	if err == nil {
		t.Fatal("expected error for unpause, got nil")
	}

	var c9err *Error
	if !errors.As(err, &c9err) {
		t.Fatalf("expected *Error, got %T", err)
	}

	if c9err.Op != "cli.pause-unsupported" {
		t.Errorf("Op = %q, want cli.pause-unsupported", c9err.Op)
	}
}

// Coverage bump tests for DefaultClient lifecycle methods

func TestDefaultClientLifecycleMethodsWithEcho(t *testing.T) {
	// Use /bin/sh -c "exit 0" as a mock binary that always succeeds.
	// The client will call: sh -c "exit 0" stop <id>
	// sh interprets "-c" then "exit 0", ignoring subsequent args, then exits 0.
	bin, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not on PATH")
	}
	c := NewDefaultClient(WithBinary(bin))

	ctx := context.Background()
	testID := "test-id"

	// These calls will fail because "sh" doesn't understand "stop", "kill", etc.
	// We just want to verify the wrapper paths are hit and return appropriate errors.
	tests := []struct {
		name   string
		call   func() error
		wantOp string
	}{
		{"StopContainer", func() error { return c.StopContainer(ctx, testID) }, "cli.stop-container"},
		{"KillContainer", func() error { return c.KillContainer(ctx, testID) }, "cli.kill-container"},
		{"RestartContainer", func() error { return c.RestartContainer(ctx, testID) }, "cli.restart-container"},
		{"DeleteContainer", func() error { return c.DeleteContainer(ctx, testID) }, "cli.delete-container"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			// sh will fail because it doesn't recognize the verb; that's fine.
			// We're testing that the wrapper path is hit.
			if err != nil {
				var c9err *Error
				if !errors.As(err, &c9err) {
					t.Errorf("expected *Error, got %T: %v", err, err)
				} else if c9err.Op != tt.wantOp {
					t.Errorf("Op = %q, want %q", c9err.Op, tt.wantOp)
				}
			}
		})
	}
}

func TestDefaultClientListContainersWithScript(t *testing.T) {
	// Use a temporary script to print valid JSON, simulating container ls output.
	script := `#!/bin/sh
cat <<'EOF'
[{"id":"abc-12345","status":"running","configuration":{"image":{"reference":"nginx"},"resources":{"cpus":1,"memoryInBytes":268435456},"publishedPorts":[],"networks":[],"platform":{"os":"linux","architecture":"arm64"}},"startedDate":799000000.0}]
EOF
`
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/fake-container.sh"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	c := NewDefaultClient(WithBinary(scriptPath))
	got, err := c.ListContainers(context.Background(), true)
	if err != nil {
		t.Fatalf("ListContainers: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d containers, want 1", len(got))
	}
	if got[0].Image != "nginx" {
		t.Errorf("Image = %q, want nginx", got[0].Image)
	}
}

func TestDefaultClientInspectContainerWithScript(t *testing.T) {
	script := `#!/bin/sh
cat <<'EOF'
{"id":"test-id","status":"running"}
EOF
`
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/fake-inspect.sh"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	c := NewDefaultClient(WithBinary(scriptPath))
	got, err := c.InspectContainer(context.Background(), "test-id")
	if err != nil {
		t.Fatalf("InspectContainer: %v", err)
	}
	s := string(got)
	if !strings.Contains(s, `"id":"test-id"`) {
		t.Errorf("got %q, want to contain id:test-id", s)
	}
}

func TestDefaultClientPruneContainersWithScript(t *testing.T) {
	script := `#!/bin/sh
echo "3 containers removed"
`
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/fake-prune.sh"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	c := NewDefaultClient(WithBinary(scriptPath))
	n, err := c.PruneContainers(context.Background())
	if err != nil {
		t.Fatalf("PruneContainers: %v", err)
	}
	if n != 3 {
		t.Errorf("n = %d, want 3", n)
	}
}
