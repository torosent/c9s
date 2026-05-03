package containers

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		name     string
		uptime   time.Duration
		expected string
	}{
		{"zero", 0, "-"},
		{"seconds only", 45 * time.Second, "45s"},
		{"minutes and seconds", 2*time.Minute + 30*time.Second, "2m30s"},
		{"hours and minutes", 2*time.Hour + 15*time.Minute, "2h15m"},
		{"days", 3*24*time.Hour + 5*time.Minute, "72h5m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatUptime(tt.uptime)
			if got != tt.expected {
				t.Errorf("formatUptime(%v) = %q, want %q", tt.uptime, got, tt.expected)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero", 0, "-"},
		{"1 KB", 1024, "1.0K"},
		{"384 MB", 384 * 1024 * 1024, "384M"},
		{"1.5 GB", 1536 * 1024 * 1024, "1.5G"},
		{"2.2 TB", 2*1024*1024*1024*1024 + 200*1024*1024*1024, "2.2T"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.bytes)
			if got != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.expected)
			}
		})
	}
}

func TestFormatShortID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			"UUID with dashes",
			"c93b06b2-0788-4779-bfb0-927c2bd6f8be",
			"c93b06b20788",
		},
		{
			"UUID without dashes",
			"c93b06b2078847799bfb0927c2bd6f8be",
			"c93b06b20788",
		},
		{
			"short ID",
			"abc123def456",
			"abc123def456",
		},
		{
			"very short ID",
			"abc",
			"abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatShortID(tt.id)
			if got != tt.expected {
				t.Errorf("formatShortID(%q) = %q, want %q", tt.id, got, tt.expected)
			}
		})
	}
}

func TestColorForState(t *testing.T) {
	p := theme.DefaultDark()

	tests := []struct {
		state    string
		expected lipgloss.Color
	}{
		{"running", p.State["running"]},
		{"exited", p.State["exited"]},
		{"paused", p.State["paused"]},
		{"unknown", p.Dim},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			got := colorForState(p, tt.state)
			if got != tt.expected {
				t.Errorf("colorForState(p, %q) = %v, want %v", tt.state, got, tt.expected)
			}
		})
	}
}
