package breadcrumbs

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Crumb represents a single breadcrumb in the navigation trail.
type Crumb struct {
	Label  string
	Screen string
}

// Trail maintains a stack of breadcrumbs for navigation.
type Trail struct {
	crumbs []Crumb
}

// New creates a new empty breadcrumb trail.
func New() *Trail {
	return &Trail{
		crumbs: make([]Crumb, 0),
	}
}

// Push adds a crumb to the trail.
func (t *Trail) Push(c Crumb) {
	t.crumbs = append(t.crumbs, c)
}

// Replace clears the trail and pushes a single root crumb. Used when switching
// to a top-level screen so navigation history doesn't accumulate (k9s only
// shows breadcrumbs for drill-downs into sub-screens / modals).
func (t *Trail) Replace(c Crumb) {
	t.crumbs = t.crumbs[:0]
	t.crumbs = append(t.crumbs, c)
}

// Clear empties the trail entirely.
func (t *Trail) Clear() {
	t.crumbs = t.crumbs[:0]
}

// Pop removes and returns the top crumb. Returns false if trail is empty.
func (t *Trail) Pop() (Crumb, bool) {
	if len(t.crumbs) == 0 {
		return Crumb{}, false
	}
	c := t.crumbs[len(t.crumbs)-1]
	t.crumbs = t.crumbs[:len(t.crumbs)-1]
	return c, true
}

// Top returns the top crumb without removing it. Returns false if trail is empty.
func (t *Trail) Top() (Crumb, bool) {
	if len(t.crumbs) == 0 {
		return Crumb{}, false
	}
	return t.crumbs[len(t.crumbs)-1], true
}

// Len returns the number of crumbs in the trail.
func (t *Trail) Len() int {
	return len(t.crumbs)
}

// Render renders the breadcrumb trail as a string, truncating in the middle if too long.
// Example: "containers > images > inspect" or "containers > … > logs"
func (t *Trail) Render(width int) string {
	if len(t.crumbs) == 0 {
		return ""
	}

	// Build full trail
	labels := make([]string, len(t.crumbs))
	for i, c := range t.crumbs {
		labels[i] = c.Label
	}
	full := strings.Join(labels, " > ")

	// If it fits, return as-is
	if lipgloss.Width(full) <= width {
		return full
	}

	// If only one crumb, truncate it
	if len(labels) == 1 {
		if width < 3 {
			return "..."
		}
		return truncateString(labels[0], width)
	}

	// Try showing first and last with ellipsis in middle
	if len(labels) >= 3 {
		short := labels[0] + " > … > " + labels[len(labels)-1]
		if lipgloss.Width(short) <= width {
			return short
		}
	}

	// Fall back to just the last crumb
	return truncateString(labels[len(labels)-1], width)
}

// truncateString truncates s to fit within maxWidth, adding "…" if needed.
func truncateString(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	if maxWidth < 3 {
		return "..."[:maxWidth]
	}
	// Roughly estimate how many runes fit
	target := maxWidth - 1 // leave room for ellipsis
	runes := []rune(s)
	if len(runes) > target {
		return string(runes[:target]) + "…"
	}
	return s[:maxWidth-1] + "…"
}
