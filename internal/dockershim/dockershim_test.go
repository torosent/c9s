package dockershim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallAndUninstall(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "docker")

	if err := Install(path, false); err != nil {
		t.Fatalf("Install: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Errorf("expected executable, got mode %v", info.Mode())
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.HasPrefix(string(body), "#!/usr/bin/env bash") {
		t.Error("shim missing shebang")
	}
	if !strings.Contains(string(body), Sentinel) {
		t.Error("shim missing sentinel header")
	}

	if err := Uninstall(path); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file removed, stat err=%v", err)
	}
}

func TestInstall_RefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "docker")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho real docker\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := Install(path, false)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected already-exists error, got %v", err)
	}
	if err := Install(path, true); err != nil {
		t.Fatalf("expected force overwrite to succeed, got %v", err)
	}
}

func TestInstall_RefusesBrokenSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "missing-target")
	link := filepath.Join(dir, "docker")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	// Sanity: target doesn't exist, so os.Stat would say IsNotExist and
	// the old (buggy) Install would proceed and write through the link.
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected target missing, got stat err=%v", err)
	}

	err := Install(link, false)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected broken-symlink to be treated as 'exists', got %v", err)
	}
	// And the symlink target should NOT have been created.
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("Install wrote through broken symlink; target now exists at %s", target)
	}
}

func TestUninstall_RefusesNonShim(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "docker")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho real docker\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := Uninstall(path)
	if err == nil || !strings.Contains(err.Error(), "does not look like a c9s shim") {
		t.Fatalf("expected non-shim refusal, got %v", err)
	}
}

func TestScript_VolumeNetworkBareSubcommand(t *testing.T) {
	// The shim must NOT pass an empty positional to `container volume` /
	// `container network` when the user types bare `docker volume`.
	// We verify by string-search rather than executing bash, so the test
	// runs everywhere.
	if !strings.Contains(Script, `"")
        # Bare 'docker volume' / 'docker network' should defer to`) {
		t.Error("Script missing empty-subcommand guard for volume/network")
	}
	if strings.Contains(Script, `*)  exec "$BIN" "$verb" "$sub" "$@" ;;`) {
		t.Error("Script still has the original two-line case that doesn't guard $sub=\"\"")
	}
}

func TestScript_InfoMapsToSystemInfo(t *testing.T) {
	if !strings.Contains(Script, `info)
    exec "$BIN" system info`) {
		t.Error("Script should map 'docker info' to 'container system info', not '--version'")
	}
}

func TestDefaultPath(t *testing.T) {
	p, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if !strings.HasSuffix(p, "/.local/bin/docker") {
		t.Errorf("DefaultPath = %q, expected ~/.local/bin/docker suffix", p)
	}
}
