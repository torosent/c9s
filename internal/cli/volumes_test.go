package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestListVolumes(t *testing.T) {
	tmp := t.TempDir()
	abs, err := filepath.Abs(filepath.Join("testdata", "volume-ls.json"))
	if err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/bash\ncat " + abs + "\n"
	scriptPath := filepath.Join(tmp, "container")
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	c := NewDefaultClient(WithBinary(scriptPath))
	vols, err := c.ListVolumes(context.Background())
	if err != nil {
		t.Fatalf("ListVolumes: %v", err)
	}
	if len(vols) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(vols))
	}
	if vols[0].Name != "api-data" {
		t.Errorf("Name = %q, want api-data", vols[0].Name)
	}
	if vols[0].SizeBytes != 104857600 {
		t.Errorf("SizeBytes = %d, want 104857600", vols[0].SizeBytes)
	}
	if len(vols[0].UsedBy) != 2 || vols[0].UsedBy[0] != "api-server" {
		t.Errorf("UsedBy = %v", vols[0].UsedBy)
	}
}

func TestListVolumesMalformedJSON(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	script := "#!/bin/bash\necho 'not-json'\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	if _, err := c.ListVolumes(context.Background()); err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestCreateVolume(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	script := "#!/bin/bash\necho \"created $3\"\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	if err := c.CreateVolume(context.Background(), "myvol"); err != nil {
		t.Fatalf("CreateVolume: %v", err)
	}
}

func TestDeleteVolume(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	script := "#!/bin/bash\necho \"deleted $3\"\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	if err := c.DeleteVolume(context.Background(), "myvol"); err != nil {
		t.Fatalf("DeleteVolume: %v", err)
	}
}

func TestPruneVolumes(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	script := "#!/bin/bash\necho '3 volumes removed'\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	count, err := c.PruneVolumes(context.Background())
	if err != nil {
		t.Fatalf("PruneVolumes: %v", err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

func TestFakeVolumes(t *testing.T) {
	f := NewFake()
	f.ListVolumesResp = []Volume{{Name: "v1"}}
	f.PruneVolumesResp = 7
	ctx := context.Background()

	vs, err := f.ListVolumes(ctx)
	if err != nil || len(vs) != 1 {
		t.Fatalf("ListVolumes returned %v, %v", vs, err)
	}
	if err := f.CreateVolume(ctx, "v1"); err != nil {
		t.Errorf("CreateVolume: %v", err)
	}
	if err := f.DeleteVolume(ctx, "v1"); err != nil {
		t.Errorf("DeleteVolume: %v", err)
	}
	n, err := f.PruneVolumes(ctx)
	if err != nil || n != 7 {
		t.Errorf("PruneVolumes = %d, %v", n, err)
	}
	want := []string{"ListVolumes", "CreateVolume", "DeleteVolume", "PruneVolumes"}
	if len(f.Calls) != len(want) {
		t.Fatalf("Calls = %v", f.Calls)
	}
	for i, c := range want {
		if f.Calls[i] != c {
			t.Errorf("Calls[%d] = %q, want %q", i, f.Calls[i], c)
		}
	}
}
