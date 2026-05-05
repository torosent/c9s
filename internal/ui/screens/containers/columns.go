package containers

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/torosent/c9s/internal/ui/theme"
)

// formatUptime formats a duration as a human-readable uptime string.
func formatUptime(uptime time.Duration) string {
	if uptime == 0 {
		return "-"
	}

	h := int(uptime.Hours())
	m := int(uptime.Minutes()) % 60
	s := int(uptime.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	} else if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// formatBytes formats bytes as a human-readable size string.
func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "-"
	}

	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1fT", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.1fG", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.0fM", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1fK", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// formatShortID extracts the first 12 hex characters from a container ID.
func formatShortID(id string) string {
	// Remove dashes
	cleaned := strings.ReplaceAll(id, "-", "")

	// Take first 12 characters
	if len(cleaned) > 12 {
		return cleaned[:12]
	}
	return cleaned
}

// colorForState returns the color for a given container state.
func colorForState(p theme.Palette, state string) color.Color {
	if c, ok := p.State[state]; ok {
		return c
	}
	return p.Dim
}
