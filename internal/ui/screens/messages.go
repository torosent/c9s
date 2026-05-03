package screens

import "github.com/torosent/c9s/internal/ui/modals"

// SuspendShellMsg requests the app suspend Bubble Tea and run a shell in the container.
type SuspendShellMsg struct {
	ID    string
	Shell string
}

// OpenModalMsg requests the app push a modal onto the stack.
type OpenModalMsg struct {
	Modal modals.Modal
}

// StatusMsg requests the app display a toast in the status bar.
type StatusMsg struct {
	Toast string
}

// PinMsg requests the app store a bookmark in the pinned store.
type PinMsg struct {
	Resource string
	ID       string
	Display  string
}
