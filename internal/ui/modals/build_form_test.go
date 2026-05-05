package modals

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestNewBuildFormPrefillsPath(t *testing.T) {
	m := NewBuildForm("./api", theme.DefaultDark())
	v := m.View(80, 30)
	if !strings.Contains(v, "./api") {
		t.Errorf("expected path prefill: %q", v)
	}
}

func TestBuildFormSubmits(t *testing.T) {
	m := NewBuildForm("./api", theme.DefaultDark())
	for _, r := range "ghcr.io/me/api:1.0" {
		m2, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = m2.(BuildFormModel)
	}
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("submit cmd nil")
	}
	gotSubmit := false
	gotClose := false
	collect(cmd, func(msg tea.Msg) {
		if r, ok := msg.(BuildSubmittedMsg); ok {
			gotSubmit = true
			if r.Opts.ContextPath != "./api" {
				t.Errorf("ContextPath=%q", r.Opts.ContextPath)
			}
			if r.Opts.Tag != "ghcr.io/me/api:1.0" {
				t.Errorf("Tag=%q", r.Opts.Tag)
			}
		}
		if _, ok := msg.(CloseModalMsg); ok {
			gotClose = true
		}
	})
	if !gotSubmit || !gotClose {
		t.Errorf("submit=%v close=%v", gotSubmit, gotClose)
	}
}

func TestBuildFormEscCancels(t *testing.T) {
	m := NewBuildForm("", theme.DefaultDark())
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc nil")
	}
	got := false
	collect(cmd, func(msg tea.Msg) {
		if _, ok := msg.(BuildCancelledMsg); ok {
			got = true
		}
	})
	if !got {
		t.Error("expected BuildCancelledMsg")
	}
}

func TestBuildFormTabCycles(t *testing.T) {
	m := NewBuildForm("", theme.DefaultDark())
	for i := 0; i < buildFieldCount+2; i++ {
		m2, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		m = m2.(BuildFormModel)
	}
	for i := 0; i < buildFieldCount+1; i++ {
		m2, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
		m = m2.(BuildFormModel)
	}
	_ = m.View(80, 30)
}

func TestBuildFormEnterAdvancesAndSubmits(t *testing.T) {
	m := NewBuildForm("./", theme.DefaultDark())
	// Three Enters advance from tag→cf→platform→submit on platform
	for i := 0; i < 3; i++ {
		m2, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		if cmd != nil && i == 2 {
			gotSubmit := false
			collect(cmd, func(msg tea.Msg) {
				if _, ok := msg.(BuildSubmittedMsg); ok {
					gotSubmit = true
				}
			})
			if !gotSubmit {
				t.Errorf("expected submit on last Enter, got cmd output without submit")
			}
		}
		m = m2.(BuildFormModel)
	}
}

func TestBuildFormTitle(t *testing.T) {
	m := NewBuildForm("", theme.DefaultDark())
	if m.Title() != "Build image" {
		t.Errorf("Title=%q", m.Title())
	}
}
