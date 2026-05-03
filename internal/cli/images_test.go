package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestListImages(t *testing.T) {
	// Test with real JSON fixture
	c := NewDefaultClient(WithBinary("cat"))
	testdata := filepath.Join("testdata", "image-ls.json")
	absPath, err := filepath.Abs(testdata)
	if err != nil {
		t.Fatal(err)
	}

	// Temporarily override binary to return fixture
	c.bin = "cat"
	originalBin := c.bin
	defer func() { c.bin = originalBin }()

	// Create a temp script that outputs the fixture
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "container")
	script := `#!/bin/bash
cat ` + absPath
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	c.bin = scriptPath
	ctx := context.Background()
	images, err := c.ListImages(ctx)
	if err != nil {
		t.Fatalf("ListImages failed: %v", err)
	}

	if len(images) != 3 {
		t.Fatalf("expected 3 images, got %d", len(images))
	}

	// Verify first image
	img := images[0]
	if img.ID != "sha256:abc123def456" {
		t.Errorf("expected ID sha256:abc123def456, got %s", img.ID)
	}
	if img.Repository != "ghcr.io/acme/api" {
		t.Errorf("expected repository ghcr.io/acme/api, got %s", img.Repository)
	}
	if img.Tag != "1.4.2" {
		t.Errorf("expected tag 1.4.2, got %s", img.Tag)
	}
	if img.Reference != "ghcr.io/acme/api:1.4.2" {
		t.Errorf("expected reference ghcr.io/acme/api:1.4.2, got %s", img.Reference)
	}
	if img.SizeBytes != 524288000 {
		t.Errorf("expected size 524288000, got %d", img.SizeBytes)
	}

	if img.ShortID != "abc123def456" {
		t.Errorf("expected short ID abc123def456, got %s", img.ShortID)
	}

	// Apple's container CLI does not emit a creation time; img.Created is
	// expected to be the zero value.
	if !img.Created.IsZero() {
		t.Errorf("expected zero created time (Apple CLI doesn't emit one), got %v", img.Created)
	}
}

func TestTagImage(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "container")
	script := `#!/bin/bash
echo "Tagged $3 as $4"
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	c := NewDefaultClient(WithBinary(scriptPath))
	ctx := context.Background()

	err := c.TagImage(ctx, "sha256:abc123", "ghcr.io/acme/api:latest")
	if err != nil {
		t.Fatalf("TagImage failed: %v", err)
	}
}

func TestDeleteImage(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "container")
	script := `#!/bin/bash
echo "Deleted image $3"
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	c := NewDefaultClient(WithBinary(scriptPath))
	ctx := context.Background()

	err := c.DeleteImage(ctx, "sha256:abc123")
	if err != nil {
		t.Fatalf("DeleteImage failed: %v", err)
	}
}

func TestPruneImages(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "container")
	script := `#!/bin/bash
echo "5 images removed"
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	c := NewDefaultClient(WithBinary(scriptPath))
	ctx := context.Background()

	count, err := c.PruneImages(ctx, false)
	if err != nil {
		t.Fatalf("PruneImages failed: %v", err)
	}

	if count != 5 {
		t.Errorf("expected 5 pruned images, got %d", count)
	}
}

func TestInspectImage(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "container")
	script := `#!/bin/bash
echo '{"id":"sha256:abc123","repository":"test"}'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	c := NewDefaultClient(WithBinary(scriptPath))
	ctx := context.Background()

	data, err := c.InspectImage(ctx, "sha256:abc123")
	if err != nil {
		t.Fatalf("InspectImage failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty inspection data")
	}
}

func TestLoadImage(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "container")
	script := `#!/bin/bash
echo "Loaded image from $4"
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	c := NewDefaultClient(WithBinary(scriptPath))
	ctx := context.Background()

	err := c.LoadImage(ctx, "/path/to/image.tar")
	if err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}
}

func TestSaveImage(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "container")
	script := `#!/bin/bash
echo "Saved image $3 to $5"
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	c := NewDefaultClient(WithBinary(scriptPath))
	ctx := context.Background()

	err := c.SaveImage(ctx, "ghcr.io/acme/api:latest", "/path/to/output.tar")
	if err != nil {
		t.Fatalf("SaveImage failed: %v", err)
	}
}

func TestFakeListImages(t *testing.T) {
	f := NewFake()
	f.ListImagesResp = []Image{
		{ID: "sha256:test", ShortID: "test", Repository: "test", Tag: "v1"},
	}

	ctx := context.Background()
	images, err := f.ListImages(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}

	if len(f.Calls) != 1 || f.Calls[0] != "ListImages" {
		t.Errorf("expected Calls=[ListImages], got %v", f.Calls)
	}
}
