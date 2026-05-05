package modals

import tea "charm.land/bubbletea/v2"

// Modal represents a temporary overlay UI (e.g., confirm dialog, help screen).
type Modal interface {
	// Init returns the initial command to run when the modal is shown.
	Init() tea.Cmd

	// Update handles incoming messages and returns the updated modal and
	// any command to execute.
	Update(msg tea.Msg) (Modal, tea.Cmd)

	// View renders the modal content at the given dimensions.
	View(width, height int) string

	// Title returns the modal's title for display.
	Title() string
}

// Stack manages a stack of modals, allowing modals to be pushed and popped.
type Stack struct {
	items []Modal
}

// Push adds a modal to the top of the stack.
func (s *Stack) Push(m Modal) {
	s.items = append(s.items, m)
}

// Pop removes and returns the top modal from the stack.
// Returns nil if the stack is empty.
func (s *Stack) Pop() Modal {
	if len(s.items) == 0 {
		return nil
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top
}

// Top returns the top modal without removing it.
// Returns nil if the stack is empty.
func (s *Stack) Top() Modal {
	if len(s.items) == 0 {
		return nil
	}
	return s.items[len(s.items)-1]
}

// Empty returns true if the stack has no modals.
func (s *Stack) Empty() bool {
	return len(s.items) == 0
}

// Len returns the number of modals in the stack.
func (s *Stack) Len() int {
	return len(s.items)
}
