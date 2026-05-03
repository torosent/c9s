package images

import (
	"testing"
	"time"
)

func TestFormatShortID(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"sha256:abc123def4567890", "abc123def456"},
		{"abc123def456", "abc123def456"},
		{"sha256:short", "short"},
		{"", ""},
	}
	for _, tt := range tests {
		got := formatShortID(tt.in)
		if got != tt.want {
			t.Errorf("formatShortID(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{0, "-"},
		{500, "500B"},
		{2048, "2.0K"},
		{1572864, "2M"},
		{1610612736, "1.5G"},
	}
	for _, tt := range tests {
		got := formatBytes(tt.in)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatCreated(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		created time.Time
		want    string
	}{
		{time.Time{}, "-"},
		{now.Add(-30 * time.Second), "30s"},
		{now.Add(-5 * time.Minute), "5m"},
		{now.Add(-3 * time.Hour), "3h"},
		{now.Add(-2 * 24 * time.Hour), "2d"},
		{now.Add(1 * time.Hour), "-"},
	}
	for _, tt := range tests {
		got := formatCreated(tt.created, now)
		if got != tt.want {
			t.Errorf("formatCreated(%v) = %q, want %q", tt.created, got, tt.want)
		}
	}
}
