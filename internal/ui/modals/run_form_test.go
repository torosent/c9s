package modals

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestNewRunFormPrefillsImage(t *testing.T) {
	m := NewRunForm("ghcr.io/acme/api:1.0", theme.DefaultDark())
	if !strings.Contains(m.View(80, 30), "ghcr.io/acme/api:1.0") {
		t.Errorf("expected image prefill in view")
	}
}

func TestRunFormDefaultsDetachTrue(t *testing.T) {
	m := NewRunForm("alpine", theme.DefaultDark())
	v := m.View(80, 30)
	// Detach default checkbox should be ticked
	if !strings.Contains(v, "[x] Detach") {
		t.Errorf("expected [x] Detach default in view: %q", v)
	}
}

func TestRunFormTabAndSpaceToggles(t *testing.T) {
	m := NewRunForm("alpine", theme.DefaultDark())
	// Tab to interactive (focus=0 name → 1 image → 2 ports → 3 env → 4 volumes → 5 interactive)
	for i := 0; i < runFieldInteractive; i++ {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
		m = m2.(RunFormModel)
	}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = m2.(RunFormModel)
	v := m.View(80, 30)
	if !strings.Contains(v, "[x] Interactive") {
		t.Errorf("expected interactive toggled: %q", v)
	}
}

func TestRunFormCtrlEnterSubmits(t *testing.T) {
	m := NewRunForm("alpine", theme.DefaultDark())
	// Move to ports field and type
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = m2.(RunFormModel)
	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = m2.(RunFormModel)
	for _, r := range "8080:80, 443:443" {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = m2.(RunFormModel)
	}
	// Ctrl-S submits as well (for terminals without Ctrl-Enter)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS, Runes: []rune("ctrl+s")})
	if cmd == nil {
		// Try the simulated Ctrl-Enter path
		_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	}
	if cmd == nil {
		t.Fatal("expected submit cmd")
	}
	gotSubmit := false
	gotClose := false
	collect(cmd, func(msg tea.Msg) {
		if r, ok := msg.(RunSubmittedMsg); ok {
			gotSubmit = true
			if r.Opts.Image != "alpine" {
				t.Errorf("Image=%q", r.Opts.Image)
			}
			if len(r.Opts.Ports) != 2 || r.Opts.Ports[0] != "8080:80" {
				t.Errorf("Ports=%v", r.Opts.Ports)
			}
			if !r.Opts.Detach {
				t.Errorf("Detach should default true")
			}
		}
		if _, ok := msg.(CloseModalMsg); ok {
			gotClose = true
		}
	})
	if !gotSubmit || !gotClose {
		t.Errorf("expected submit + close, got submit=%v close=%v", gotSubmit, gotClose)
	}
}

func TestRunFormEscCancels(t *testing.T) {
	m := NewRunForm("", theme.DefaultDark())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc nil")
	}
	got := false
	collect(cmd, func(msg tea.Msg) {
		if _, ok := msg.(RunCancelledMsg); ok {
			got = true
		}
	})
	if !got {
		t.Error("expected RunCancelledMsg")
	}
}

func TestRunFormShiftTabCycles(t *testing.T) {
	m := NewRunForm("alpine", theme.DefaultDark())
	for i := 0; i < runFieldCount+1; i++ {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		m = m2.(RunFormModel)
	}
	_ = m.View(80, 30)
}

func TestRunFormSubmitMapsAllFieldsToCLIRunOpts(t *testing.T) {
	m := NewRunForm("img", theme.DefaultDark())
	// Verify form value mapping by setting them through the public Update path
	for _, r := range "myname" {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = m2.(RunFormModel)
	}
	// Tab forward to volumes field
	for i := 0; i < runFieldVolumes; i++ {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
		m = m2.(RunFormModel)
	}
	for _, r := range "data:/data" {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = m2.(RunFormModel)
	}
	// Ctrl-D submit
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if cmd == nil {
		t.Fatal("submit cmd nil")
	}
	var got cli.RunOpts
	collect(cmd, func(msg tea.Msg) {
		if r, ok := msg.(RunSubmittedMsg); ok {
			got = r.Opts
		}
	})
	if got.Name != "myname" {
		t.Errorf("Name=%q", got.Name)
	}
	if got.Image != "img" {
		t.Errorf("Image=%q", got.Image)
	}
	if len(got.Volumes) != 1 || got.Volumes[0] != "data:/data" {
		t.Errorf("Volumes=%v", got.Volumes)
	}
}

func TestSplitCSV(t *testing.T) {
	if v := splitCSV(""); v != nil {
		t.Errorf("got %v", v)
	}
	if v := splitCSV("a, b ,c,, "); len(v) != 3 || v[2] != "c" {
		t.Errorf("got %v", v)
	}
}
