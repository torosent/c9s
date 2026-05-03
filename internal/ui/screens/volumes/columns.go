package volumes

import (
	"encoding/json"
	"fmt"
)

// formatBytes formats bytes as a short human-readable size string.
func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "-"
	}
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
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

// jsonOf returns the JSON representation of a value, suitable for the
// inspect modal. Used by the volumes/networks/builder screens which do
// not have a dedicated `inspect` CLI subcommand.
func jsonOf(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
