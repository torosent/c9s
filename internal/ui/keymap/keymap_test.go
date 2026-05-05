package keymap

import (
	"sort"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestDefaultHasExpectedBindings(t *testing.T) {
	m := Default()

	expected := []string{
		"quit", "help", "filter", "palette", "mark", "mark_all",
		"refresh", "up", "down", "top", "bottom", "escape",
		"header_toggle", "interrupt",
	}

	for _, name := range expected {
		if _, ok := m.Get(name); !ok {
			t.Errorf("expected binding %q in Default(), not found", name)
		}
	}
}

func TestMatchesQuit(t *testing.T) {
	m := Default()

	msg := tea.KeyPressMsg{Code: 'q', Text: "q"}
	if !m.Matches("quit", msg) {
		t.Error("expected 'q' to match 'quit'")
	}
}

func TestMatchesInterrupt(t *testing.T) {
	m := Default()

	msg := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	if !m.Matches("interrupt", msg) {
		t.Error("expected ctrl+c to match 'interrupt'")
	}
}

func TestMatchesHeaderToggle(t *testing.T) {
	m := Default()

	msg := tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl}
	if !m.Matches("header_toggle", msg) {
		t.Error("expected ctrl+e to match 'header_toggle'")
	}
}

func TestMatchesEscape(t *testing.T) {
	m := Default()

	msg := tea.KeyPressMsg{Code: tea.KeyEsc}
	if !m.Matches("escape", msg) {
		t.Error("expected esc to match 'escape'")
	}
}

func TestMatchesUpVimAlias(t *testing.T) {
	m := Default()

	msg := tea.KeyPressMsg{Code: 'k', Text: "k"}
	if !m.Matches("up", msg) {
		t.Error("expected 'k' to match 'up'")
	}
}

func TestMatchesUpArrow(t *testing.T) {
	m := Default()

	msg := tea.KeyPressMsg{Code: tea.KeyUp}
	if !m.Matches("up", msg) {
		t.Error("expected arrow up to match 'up'")
	}
}

func TestOverrideBinding(t *testing.T) {
	m := Default()

	// Override quit to use 'Q' instead of 'q'
	m.Add("quit", Binding{Keys: []string{"Q"}})

	// 'q' should no longer match
	msgLowerQ := tea.KeyPressMsg{Code: 'q', Text: "q"}
	if m.Matches("quit", msgLowerQ) {
		t.Error("expected 'q' to NOT match 'quit' after override")
	}

	// 'Q' should match
	msgUpperQ := tea.KeyPressMsg{Code: 'Q', Text: "Q"}
	if !m.Matches("quit", msgUpperQ) {
		t.Error("expected 'Q' to match 'quit' after override")
	}
}

func TestNames(t *testing.T) {
	m := &Map{entries: make(map[string]Binding)}
	m.Add("zebra", Binding{Keys: []string{"z"}})
	m.Add("apple", Binding{Keys: []string{"a"}})
	m.Add("banana", Binding{Keys: []string{"b"}})

	names := m.Names()

	// Should be sorted
	expected := []string{"apple", "banana", "zebra"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(names))
	}

	if !sort.StringsAreSorted(names) {
		t.Error("expected names to be sorted")
	}

	for i, exp := range expected {
		if names[i] != exp {
			t.Errorf("names[%d] = %q, want %q", i, names[i], exp)
		}
	}
}

func TestApply_SingleOverride(t *testing.T) {
	m := Default()

	overrides := map[string]string{
		"quit": "Q",
	}

	m = Apply(m, overrides)

	// 'q' should no longer match
	msgLowerQ := tea.KeyPressMsg{Code: 'q', Text: "q"}
	if m.Matches("quit", msgLowerQ) {
		t.Error("expected 'q' to NOT match 'quit' after override")
	}

	// 'Q' should match
	msgUpperQ := tea.KeyPressMsg{Code: 'Q', Text: "Q"}
	if !m.Matches("quit", msgUpperQ) {
		t.Error("expected 'Q' to match 'quit' after override")
	}
}

func TestApply_MultipleOverrides(t *testing.T) {
	m := Default()

	overrides := map[string]string{
		"quit":   "x",
		"filter": "f",
		"mark":   "m",
	}

	m = Apply(m, overrides)

	// Check quit -> x
	msgX := tea.KeyPressMsg{Code: 'x', Text: "x"}
	if !m.Matches("quit", msgX) {
		t.Error("expected 'x' to match 'quit'")
	}

	// Check filter -> f
	msgF := tea.KeyPressMsg{Code: 'f', Text: "f"}
	if !m.Matches("filter", msgF) {
		t.Error("expected 'f' to match 'filter'")
	}

	// Check mark -> m
	msgM := tea.KeyPressMsg{Code: 'm', Text: "m"}
	if !m.Matches("mark", msgM) {
		t.Error("expected 'm' to match 'mark'")
	}

	// Check that unmodified bindings still work
	msgHelp := tea.KeyPressMsg{Code: '?', Text: "?"}
	if !m.Matches("help", msgHelp) {
		t.Error("expected '?' to still match 'help'")
	}
}

func TestApply_EmptyOverrides(t *testing.T) {
	m := Default()
	original := m

	m = Apply(m, map[string]string{})

	// Should be unchanged
	if len(m.entries) != len(original.entries) {
		t.Error("expected no changes with empty overrides")
	}
}

func TestApply_NonexistentBinding(t *testing.T) {
	m := Default()

	overrides := map[string]string{
		"nonexistent": "z",
	}

	// Should not panic, just ignore
	_ = Apply(m, overrides)
}

func TestApply_PreservesHelpDescription(t *testing.T) {
	m := Default()
	originalBinding, _ := m.Get("quit")

	overrides := map[string]string{
		"quit": "Q",
	}

	m = Apply(m, overrides)

	newBinding, ok := m.Get("quit")
	if !ok {
		t.Fatal("quit binding not found after apply")
	}

	// Help and Description should be preserved
	if newBinding.Help != originalBinding.Help {
		t.Errorf("expected Help to be preserved, got %q, want %q", newBinding.Help, originalBinding.Help)
	}
	if newBinding.Description != originalBinding.Description {
		t.Errorf("expected Description to be preserved, got %q, want %q", newBinding.Description, originalBinding.Description)
	}
}
