package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/ui/modals"
)

// progressModalWrap adapts modals.ProgressModel (a tea.Model with a
// no-arg View) into the modals.Modal interface so it can live on the
// app's modal stack.
type progressModalWrap struct {
	p *modals.ProgressModel
}

// Init implements modals.Modal.
func (w progressModalWrap) Init() tea.Cmd {
	if w.p == nil {
		return nil
	}
	return w.p.Init()
}

// Update implements modals.Modal.
func (w progressModalWrap) Update(msg tea.Msg) (modals.Modal, tea.Cmd) {
	if w.p == nil {
		return w, nil
	}
	updated, cmd := w.p.Update(msg)
	if pm, ok := updated.(*modals.ProgressModel); ok {
		w.p = pm
	}
	return w, cmd
}

// View implements modals.Modal.
func (w progressModalWrap) View(width, height int) string {
	if w.p == nil {
		return ""
	}
	return w.p.View()
}

// Title implements modals.Modal.
func (w progressModalWrap) Title() string {
	return "Progress"
}
