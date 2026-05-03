package ui

import (
	"regexp"
	"strings"
	"testing"

	"github.com/torosent/c9s/internal/ui/theme"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

func stripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }
func runeWidth(s string) int    { return len([]rune(s)) }

func TestStatusBarRendersScreenAndSummary(t *testing.T) {
	sb := NewStatusBar(theme.DefaultDark()).Update(StatusUpdate{
		Screen:  "containers",
		Summary: "16 items · 14 running, 2 exited",
		Hint:    "⏎ inspect",
	})
	out := stripANSI(sb.View(120, false))
	for _, want := range []string{"containers", "16 items", "⏎ inspect"} {
		if !strings.Contains(out, want) {
			t.Errorf("View missing %q: %q", want, out)
		}
	}
}

func TestStatusBarToastOverridesHint(t *testing.T) {
	sb := NewStatusBar(theme.DefaultDark()).Update(StatusUpdate{
		Screen: "containers",
		Hint:   "⏎ inspect",
		Toast:  "stop failed: no such container",
	})
	out := stripANSI(sb.View(120, false))
	if !strings.Contains(out, "stop failed") {
		t.Errorf("Toast not rendered: %q", out)
	}
	if strings.Contains(out, "⏎ inspect") {
		t.Errorf("Hint should be replaced by toast, got: %q", out)
	}
}

func TestStatusBarTrimsToWidth(t *testing.T) {
	sb := NewStatusBar(theme.DefaultDark()).Update(StatusUpdate{
		Screen:  "containers",
		Summary: strings.Repeat("x", 500),
	})
	out := stripANSI(sb.View(40, false))
	if got := runeWidth(out); got > 40 {
		t.Errorf("View exceeded width: %d > 40 (%q)", got, out)
	}
}
