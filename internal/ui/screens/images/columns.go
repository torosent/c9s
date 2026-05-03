// Package images implements the :images resource screen.
package images

import (
	"fmt"
	"strings"
	"time"
)

// formatShortID strips the sha256: prefix and returns the first 12 hex chars.
func formatShortID(id string) string {
	cleaned := strings.TrimPrefix(id, "sha256:")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	if len(cleaned) > 12 {
		return cleaned[:12]
	}
	return cleaned
}

// formatBytes formats bytes as a short human-readable size string.
func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "-"
	}

	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
		tb = gb * 1024
	)

	switch {
	case bytes >= tb:
		return fmt.Sprintf("%.1fT", float64(bytes)/tb)
	case bytes >= gb:
		return fmt.Sprintf("%.1fG", float64(bytes)/gb)
	case bytes >= mb:
		return fmt.Sprintf("%.0fM", float64(bytes)/mb)
	case bytes >= kb:
		return fmt.Sprintf("%.1fK", float64(bytes)/kb)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// formatCreated returns a short relative-time string like "3d", "2h", "1m".
func formatCreated(created, now time.Time) string {
	if created.IsZero() {
		return "-"
	}
	d := now.Sub(created)
	if d < 0 {
		return "-"
	}
	switch {
	case d >= 24*time.Hour:
		days := int(d / (24 * time.Hour))
		return fmt.Sprintf("%dd", days)
	case d >= time.Hour:
		return fmt.Sprintf("%dh", int(d/time.Hour))
	case d >= time.Minute:
		return fmt.Sprintf("%dm", int(d/time.Minute))
	default:
		return fmt.Sprintf("%ds", int(d/time.Second))
	}
}
