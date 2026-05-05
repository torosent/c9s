package modals

import tea "charm.land/bubbletea/v2"

// CloseModalMsg signals that the current modal should be closed.
type CloseModalMsg struct{}

// ConfirmResult carries the user's decision from a confirm dialog.
type ConfirmResult struct {
	Confirmed bool
	Tag       string
}

// ConfirmResultMsg is emitted when a confirm modal is resolved.
type ConfirmResultMsg struct {
	Result ConfirmResult
}

// StatusMsg is a status bar message emitted by modals.
type StatusMsg string

// CloseModal returns a command that closes the current modal.
func CloseModal() func() tea.Msg {
	return func() tea.Msg { return CloseModalMsg{} }
}
