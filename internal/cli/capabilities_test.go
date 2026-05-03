package cli

import (
	"context"
	"errors"
	"testing"
)

func TestProbeParsesVersion(t *testing.T) {
	f := &Fake{VersionResp: "container CLI version 0.12.1 (build: release, commit: e9891b3)"}
	caps, err := ProbeCapabilities(context.Background(), f)
	if err != nil {
		t.Fatalf("ProbeCapabilities: %v", err)
	}
	if caps.Major != 0 || caps.Minor != 12 || caps.Patch != 1 {
		t.Errorf("got %d.%d.%d, want 0.12.1", caps.Major, caps.Minor, caps.Patch)
	}
	if caps.Version != "0.12.1" {
		t.Errorf("Version = %q, want 0.12.1", caps.Version)
	}
}

func TestProbeUnparseable(t *testing.T) {
	f := &Fake{VersionResp: "weird unparseable output"}
	caps, err := ProbeCapabilities(context.Background(), f)
	if err != nil {
		t.Fatalf("ProbeCapabilities: %v", err)
	}
	if caps.Major != 0 || caps.Minor != 0 || caps.Patch != 0 {
		t.Errorf("expected zeros, got %d.%d.%d", caps.Major, caps.Minor, caps.Patch)
	}
	if caps.Version == "" {
		t.Errorf("Version should preserve raw on unparseable output")
	}
}

func TestProbePropagatesErr(t *testing.T) {
	sentinel := errors.New("services not running")
	f := &Fake{VersionErr: sentinel}
	if _, err := ProbeCapabilities(context.Background(), f); !errors.Is(err, sentinel) {
		t.Errorf("expected wrapped sentinel, got %v", err)
	}
}
