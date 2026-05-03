package cli

import (
	"os"
	"testing"
	"time"
)

func TestContainerJSONFromFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/list.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	// Use a fixed time for testing uptime calculations
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)

	containers, err := parseContainers(raw, now)
	if err != nil {
		t.Fatalf("parseContainers failed: %v", err)
	}

	if len(containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(containers))
	}

	// Test stopped container (first in fixture)
	stopped := containers[0]
	if stopped.ID != "c93b06b2-0788-4779-bfb0-927c2bd6f8be" {
		t.Errorf("expected ID 'c93b06b2-0788-4779-bfb0-927c2bd6f8be', got '%s'", stopped.ID)
	}

	expectedShortID := "c93b06b20788"
	if stopped.ShortID != expectedShortID {
		t.Errorf("expected ShortID '%s', got '%s'", expectedShortID, stopped.ShortID)
	}

	if stopped.Image != "mcr.microsoft.com/dts/dts-emulator:latest" {
		t.Errorf("expected Image 'mcr.microsoft.com/dts/dts-emulator:latest', got '%s'", stopped.Image)
	}

	if stopped.Status != "stopped" {
		t.Errorf("expected Status 'stopped', got '%s'", stopped.Status)
	}

	if stopped.Uptime != 0 {
		t.Errorf("expected stopped container to have Uptime=0, got %v", stopped.Uptime)
	}

	if stopped.OS != "linux" {
		t.Errorf("expected OS 'linux', got '%s'", stopped.OS)
	}

	if stopped.Arch != "arm64" {
		t.Errorf("expected Arch 'arm64', got '%s'", stopped.Arch)
	}

	if stopped.CPU != 4 {
		t.Errorf("expected CPU 4, got %d", stopped.CPU)
	}

	if stopped.MemBytes != 1073741824 {
		t.Errorf("expected MemBytes 1073741824, got %d", stopped.MemBytes)
	}

	if len(stopped.Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(stopped.Ports))
	} else {
		if stopped.Ports[0].Proto != "tcp" || stopped.Ports[0].HostPort != 8080 || stopped.Ports[0].ContainerPort != 8080 {
			t.Errorf("port[0] mismatch: %+v", stopped.Ports[0])
		}
	}

	if len(stopped.Networks) != 1 {
		t.Errorf("expected 1 network, got %d", len(stopped.Networks))
	} else {
		if stopped.Networks[0].Network != "default" || stopped.Networks[0].Hostname != "test-container" {
			t.Errorf("network[0] mismatch: %+v", stopped.Networks[0])
		}
	}

	// Test running container (second in fixture)
	running := containers[1]
	if running.ID != "a12b34c5-6789-4def-abc0-123456789abc" {
		t.Errorf("expected ID 'a12b34c5-6789-4def-abc0-123456789abc', got '%s'", running.ID)
	}

	if running.Status != "running" {
		t.Errorf("expected Status 'running', got '%s'", running.Status)
	}

	if running.Uptime == 0 {
		t.Errorf("expected running container to have non-zero Uptime, got 0")
	}

	// Verify StartedAt is converted correctly from Apple epoch
	// startedDate: 799200000.0 seconds from 2001-01-01 UTC
	appleEpoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedStarted := appleEpoch.Add(time.Duration(799200000.0 * float64(time.Second)))
	if !running.StartedAt.Equal(expectedStarted) {
		t.Errorf("expected StartedAt %v, got %v", expectedStarted, running.StartedAt)
	}

	// Verify uptime calculation
	expectedUptime := now.Sub(running.StartedAt)
	if running.Uptime != expectedUptime {
		t.Errorf("expected Uptime %v, got %v", expectedUptime, running.Uptime)
	}
}

func TestParseListEmpty(t *testing.T) {
	raw := []byte("[]")
	now := time.Now()

	containers, err := parseContainers(raw, now)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(containers) != 0 {
		t.Errorf("expected empty slice, got %d items", len(containers))
	}
}

func TestParseInvalidJSON(t *testing.T) {
	raw := []byte("not json")
	now := time.Now()

	_, err := parseContainers(raw, now)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	cliErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	if cliErr.Op != "cli.parse-containers" {
		t.Errorf("expected Op='cli.parse-containers', got '%s'", cliErr.Op)
	}
}

func TestParsePruneCount(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"3 containers removed", 3},
		{"1 container removed", 1},
		{"", 0},
		{"no containers to remove", 0},
		{"42 containers removed", 42},
	}
	for _, tc := range cases {
		if got := parsePruneCount([]byte(tc.in)); got != tc.want {
			t.Errorf("parsePruneCount(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
