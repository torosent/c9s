package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestListRegistries(t *testing.T) {
	abs, _ := filepath.Abs(filepath.Join("testdata", "registry-list.json"))
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	script := "#!/bin/bash\ncat " + abs + "\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	regs, err := c.ListRegistries(context.Background())
	if err != nil {
		t.Fatalf("ListRegistries: %v", err)
	}
	if len(regs) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(regs))
	}
	if regs[0].Host != "ghcr.io" || !regs[0].Default {
		t.Errorf("got %+v", regs[0])
	}
}

func TestListRegistriesNull(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho 'null'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	regs, err := c.ListRegistries(context.Background())
	if err != nil {
		t.Fatalf("ListRegistries null: %v", err)
	}
	if len(regs) != 0 {
		t.Errorf("expected 0 entries, got %d", len(regs))
	}
}

func TestRegistryLoginPasswordViaStdin(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	stdinFile := filepath.Join(tmp, "stdin.txt")
	argsFile := filepath.Join(tmp, "args.txt")
	script := "#!/bin/bash\necho \"$@\" > " + argsFile + "\ncat > " + stdinFile + "\nexit 0\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	if err := c.RegistryLogin(context.Background(), "ghcr.io", "alice", "topsecret"); err != nil {
		t.Fatalf("RegistryLogin: %v", err)
	}

	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(args); !contains(got, "--password-stdin") {
		t.Errorf("expected --password-stdin in args, got %q", got)
	}
	if got := string(args); contains(got, "topsecret") {
		t.Errorf("password should not appear in args, got %q", got)
	}

	stdin, err := os.ReadFile(stdinFile)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(stdin); !contains(got, "topsecret") {
		t.Errorf("expected stdin to contain password, got %q", got)
	}
}

func TestRegistryLogout(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	if err := c.RegistryLogout(context.Background(), "ghcr.io"); err != nil {
		t.Errorf("RegistryLogout: %v", err)
	}
}

func TestRegistrySetDefault(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	if err := c.RegistrySetDefault(context.Background(), "ghcr.io"); err != nil {
		t.Errorf("RegistrySetDefault: %v", err)
	}
}

func TestFakeRegistry(t *testing.T) {
	f := NewFake()
	f.ListRegistriesResp = []RegistryEntry{{Host: "x"}}
	ctx := context.Background()
	if _, err := f.ListRegistries(ctx); err != nil {
		t.Fatal(err)
	}
	if err := f.RegistryLogin(ctx, "x", "u", "p"); err != nil {
		t.Errorf("RegistryLogin: %v", err)
	}
	if err := f.RegistryLogout(ctx, "x"); err != nil {
		t.Errorf("RegistryLogout: %v", err)
	}
	if err := f.RegistrySetDefault(ctx, "x"); err != nil {
		t.Errorf("RegistrySetDefault: %v", err)
	}
	want := []string{"ListRegistries", "RegistryLogin", "RegistryLogout", "RegistrySetDefault"}
	if len(f.Calls) != len(want) {
		t.Fatalf("Calls = %v", f.Calls)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
