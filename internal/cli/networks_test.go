package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestListNetworks(t *testing.T) {
	abs, _ := filepath.Abs(filepath.Join("testdata", "network-ls.json"))
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	script := "#!/bin/bash\ncat " + abs + "\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	c := NewDefaultClient(WithBinary(scriptPath))
	ns, err := c.ListNetworks(context.Background())
	if err != nil {
		t.Fatalf("ListNetworks: %v", err)
	}
	if len(ns) != 2 {
		t.Fatalf("expected 2 networks, got %d", len(ns))
	}
	if ns[0].Name != "bridge0" || ns[0].Subnet != "192.168.64.0/24" {
		t.Errorf("got %+v", ns[0])
	}
}

func TestListNetworksEmpty(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	script := "#!/bin/bash\necho 'null'\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	ns, err := c.ListNetworks(context.Background())
	if err != nil {
		t.Fatalf("ListNetworks empty: %v", err)
	}
	if len(ns) != 0 {
		t.Errorf("expected 0 networks, got %d", len(ns))
	}
}

func TestCreateAndDeleteNetwork(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	ctx := context.Background()
	if err := c.CreateNetwork(ctx, "foo"); err != nil {
		t.Errorf("CreateNetwork: %v", err)
	}
	if err := c.DeleteNetwork(ctx, "foo"); err != nil {
		t.Errorf("DeleteNetwork: %v", err)
	}
}

func TestPruneNetworks(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho '2 networks removed'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	n, err := c.PruneNetworks(context.Background())
	if err != nil {
		t.Fatalf("PruneNetworks: %v", err)
	}
	if n != 2 {
		t.Errorf("count = %d, want 2", n)
	}
}

func TestFakeNetworks(t *testing.T) {
	f := NewFake()
	f.ListNetworksResp = []Network{{Name: "n1"}}
	f.PruneNetworksResp = 4
	ctx := context.Background()
	if _, err := f.ListNetworks(ctx); err != nil {
		t.Errorf("ListNetworks: %v", err)
	}
	if err := f.CreateNetwork(ctx, "x"); err != nil {
		t.Errorf("CreateNetwork: %v", err)
	}
	if err := f.DeleteNetwork(ctx, "x"); err != nil {
		t.Errorf("DeleteNetwork: %v", err)
	}
	n, _ := f.PruneNetworks(ctx)
	if n != 4 {
		t.Errorf("PruneNetworks = %d, want 4", n)
	}
}
