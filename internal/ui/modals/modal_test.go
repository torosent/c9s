package modals

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// dummyModal is a stub implementation for testing.
type dummyModal struct {
	title string
}

func (d *dummyModal) Init() tea.Cmd {
	return nil
}

func (d *dummyModal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	return d, nil
}

func (d *dummyModal) View(width, height int) string {
	return "dummy modal"
}

func (d *dummyModal) Title() string {
	return d.title
}

// Compile-time assertion that dummyModal implements Modal.
var _ Modal = (*dummyModal)(nil)

func TestStackStartsEmpty(t *testing.T) {
	s := &Stack{}
	if !s.Empty() {
		t.Error("expected new Stack to be empty")
	}
	if s.Len() != 0 {
		t.Errorf("expected Len() = 0, got %d", s.Len())
	}
}

func TestStackPushPop(t *testing.T) {
	s := &Stack{}
	m := &dummyModal{title: "test"}

	s.Push(m)

	if s.Empty() {
		t.Error("expected Stack to not be empty after Push")
	}

	if s.Len() != 1 {
		t.Errorf("expected Len() = 1, got %d", s.Len())
	}

	top := s.Top()
	if top != m {
		t.Error("expected Top() to return pushed modal")
	}

	popped := s.Pop()
	if popped != m {
		t.Error("expected Pop() to return pushed modal")
	}

	if !s.Empty() {
		t.Error("expected Stack to be empty after Pop")
	}
}

func TestStackPopEmpty(t *testing.T) {
	s := &Stack{}
	popped := s.Pop()
	if popped != nil {
		t.Error("expected Pop() on empty stack to return nil")
	}
}

func TestStackLIFO(t *testing.T) {
	s := &Stack{}
	m1 := &dummyModal{title: "first"}
	m2 := &dummyModal{title: "second"}

	s.Push(m1)
	s.Push(m2)

	if s.Len() != 2 {
		t.Errorf("expected Len() = 2, got %d", s.Len())
	}

	// Top should return m2 (last pushed)
	if s.Top() != m2 {
		t.Error("expected Top() to return second modal")
	}

	// First pop should return m2
	popped1 := s.Pop()
	if popped1 != m2 {
		t.Error("expected first Pop() to return second modal")
	}

	// Second pop should return m1
	popped2 := s.Pop()
	if popped2 != m1 {
		t.Error("expected second Pop() to return first modal")
	}

	if !s.Empty() {
		t.Error("expected Stack to be empty after popping both")
	}
}
