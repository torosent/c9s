package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBuilderStatus(t *testing.T) {
	abs, _ := filepath.Abs(filepath.Join("testdata", "builder-status.json"))
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	script := "#!/bin/bash\ncat " + abs + "\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	st, err := c.BuilderStatus(context.Background())
	if err != nil {
		t.Fatalf("BuilderStatus: %v", err)
	}
	if st.State != "running" || st.CPUs != 4 {
		t.Errorf("got %+v", st)
	}
	if st.MemoryBytes != 8589934592 {
		t.Errorf("MemoryBytes = %d", st.MemoryBytes)
	}
}

func TestBuilderStatusNull(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho 'null'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	st, err := c.BuilderStatus(context.Background())
	if err != nil {
		t.Fatalf("BuilderStatus null: %v", err)
	}
	if st.State != "unknown" {
		t.Errorf("State = %q, want unknown", st.State)
	}
}

func TestBuilderLifecycle(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	ctx := context.Background()
	if err := c.BuilderStart(ctx); err != nil {
		t.Errorf("BuilderStart: %v", err)
	}
	if err := c.BuilderStop(ctx); err != nil {
		t.Errorf("BuilderStop: %v", err)
	}
	if err := c.BuilderDelete(ctx); err != nil {
		t.Errorf("BuilderDelete: %v", err)
	}
}

func TestFakeBuilder(t *testing.T) {
	f := NewFake()
	f.BuilderStatusResp = BuilderStatus{State: "running", CPUs: 2}
	ctx := context.Background()
	st, _ := f.BuilderStatus(ctx)
	if st.State != "running" {
		t.Errorf("State = %q", st.State)
	}
	_ = f.BuilderStart(ctx)
	_ = f.BuilderStop(ctx)
	_ = f.BuilderDelete(ctx)
	want := []string{"BuilderStatus", "BuilderStart", "BuilderStop", "BuilderDelete"}
	if len(f.Calls) != len(want) {
		t.Fatalf("Calls = %v", f.Calls)
	}
	for i, c := range want {
		if f.Calls[i] != c {
			t.Errorf("Calls[%d] = %q, want %q", i, f.Calls[i], c)
		}
	}
}
