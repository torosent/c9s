package modals

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestNewSortPicker(t *testing.T) {
	cols := []SortColumn{
		{Key: "name", Label: "Name"},
		{Key: "age", Label: "Age"},
	}
	m := NewSortPicker(cols, theme.DefaultDark())

	view := m.View(80, 24)
	if !strings.Contains(view, "Name") {
		t.Errorf("expected view to contain 'Name', got: %s", view)
	}
	if !strings.Contains(view, "Age") {
		t.Errorf("expected view to contain 'Age', got: %s", view)
	}
	if !strings.Contains(view, "ascending") {
		t.Errorf("expected default direction 'ascending', got: %s", view)
	}
}

func TestSortPicker_ReverseToggle(t *testing.T) {
	cols := []SortColumn{{Key: "name", Label: "Name"}}
	m := NewSortPicker(cols, theme.DefaultDark())

	// Press 'r' to toggle reverse
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updatedModel.(SortPickerModel)

	view := m.View(80, 24)
	if !strings.Contains(view, "descending") {
		t.Errorf("expected 'descending' after toggle, got: %s", view)
	}

	// Press 'r' again to toggle back
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updatedModel.(SortPickerModel)

	view = m.View(80, 24)
	if !strings.Contains(view, "ascending") {
		t.Errorf("expected 'ascending' after second toggle, got: %s", view)
	}
}

func TestSortPicker_EnterEmitsSortPickedMsg(t *testing.T) {
	cols := []SortColumn{{Key: "name", Label: "Name"}}
	m := NewSortPicker(cols, theme.DefaultDark())

	// Press enter
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected cmd after pressing enter")
	}

	msg := cmd()
	picked, ok := msg.(SortPickedMsg)
	if !ok {
		t.Fatalf("expected SortPickedMsg, got %T", msg)
	}
	if picked.Key != "name" {
		t.Errorf("expected key 'name', got %s", picked.Key)
	}
	if picked.Reverse {
		t.Errorf("expected Reverse=false by default")
	}
}

func TestSortPicker_EnterWithReverse(t *testing.T) {
	cols := []SortColumn{{Key: "age", Label: "Age"}}
	m := NewSortPicker(cols, theme.DefaultDark())

	// Toggle reverse first
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updatedModel.(SortPickerModel)

	// Press enter
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msg := cmd()
	picked, ok := msg.(SortPickedMsg)
	if !ok {
		t.Fatalf("expected SortPickedMsg, got %T", msg)
	}
	if !picked.Reverse {
		t.Errorf("expected Reverse=true after toggle")
	}
}

func TestSortPicker_EscEmitsCloseModalMsg(t *testing.T) {
	cols := []SortColumn{{Key: "name", Label: "Name"}}
	m := NewSortPicker(cols, theme.DefaultDark())

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected cmd after pressing esc")
	}

	msg := cmd()
	if _, ok := msg.(CloseModalMsg); !ok {
		t.Fatalf("expected CloseModalMsg, got %T", msg)
	}
}

func TestSortPicker_Title(t *testing.T) {
	cols := []SortColumn{{Key: "name", Label: "Name"}}
	m := NewSortPicker(cols, theme.DefaultDark())

	if m.Title() != "Sort" {
		t.Errorf("expected title 'Sort', got %s", m.Title())
	}
}
