package modals

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/torosent/c9s/internal/ui/theme"
)

// TextInputResult carries the user's submission from a TextInputModel.
type TextInputResult struct {
	Label string
	Value string
}

// TextInputResultMsg is emitted when the user presses Enter in a TextInputModel.
type TextInputResultMsg struct {
	Result TextInputResult
}

// TextInputCancelledMsg is emitted when the user presses Esc.
type TextInputCancelledMsg struct {
	Label string
}

// TextInputValidator is an optional function that screens can pass to
// validate input. Returning a non-empty error string blocks submission and
// shows the message in the modal footer.
type TextInputValidator func(value string) string

// TextInputModel is a single-field prompt modal. Each caller subscribes to
// results via the unique Label they provide.
type TextInputModel struct {
	label     string
	prompt    string
	field     textinput.Model
	validator TextInputValidator
	errMsg    string
	palette   theme.Palette
}

// NewTextInput creates a new prompt modal. The label is echoed back on
// submission so callers can route results.
func NewTextInput(label, prompt, initial string, p theme.Palette) TextInputModel {
	field := textinput.New()
	field.Placeholder = ""
	field.Prompt = "> "
	field.CharLimit = 256
	field.Width = 60
	styleTextInput(&field, p)
	if initial != "" {
		field.SetValue(initial)
	}
	field.Focus()
	return TextInputModel{
		label:   label,
		prompt:  prompt,
		field:   field,
		palette: p,
	}
}

// WithValidator returns a copy of the modal with the given validator attached.
func (m TextInputModel) WithValidator(v TextInputValidator) TextInputModel {
	m.validator = v
	return m
}

// Init implements Modal.
func (m TextInputModel) Init() tea.Cmd { return textinput.Blink }

// Update implements Modal.
func (m TextInputModel) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			label := m.label
			return m, tea.Batch(
				func() tea.Msg { return TextInputCancelledMsg{Label: label} },
				CloseModal(),
			)
		case tea.KeyEnter:
			value := m.field.Value()
			if m.validator != nil {
				if errMsg := m.validator(value); errMsg != "" {
					m.errMsg = errMsg
					return m, nil
				}
			}
			label := m.label
			return m, tea.Batch(
				func() tea.Msg {
					return TextInputResultMsg{Result: TextInputResult{
						Label: label, Value: strings.TrimSpace(value),
					}}
				},
				CloseModal(),
			)
		}
	}
	var cmd tea.Cmd
	m.field, cmd = m.field.Update(msg)
	m.errMsg = ""
	return m, cmd
}

// View implements Modal.
func (m TextInputModel) View(width, height int) string {
	const boxW = 70
	innerW := boxW - 6
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Border).
		BorderBackground(m.palette.Bg).
		Background(m.palette.Bg).
		Foreground(m.palette.Fg).
		Padding(1, 2).
		Width(boxW)

	bg := lipgloss.NewStyle().Foreground(m.palette.Fg).Background(m.palette.Bg)
	prompt := lipgloss.NewStyle().Foreground(m.palette.Fg).Background(m.palette.Bg).Render(m.prompt)
	body := bg.Width(innerW).Render(prompt) + "\n" +
		bg.Width(innerW).Render(" ") + "\n" +
		renderTextInput(m.field, m.palette, innerW) + "\n"
	if m.errMsg != "" {
		errLine := lipgloss.NewStyle().Foreground(m.palette.Error).Background(m.palette.Bg).Render(m.errMsg)
		body += bg.Width(innerW).Render(" ") + "\n" + bg.Width(innerW).Render(errLine) + "\n"
	}
	help := lipgloss.NewStyle().Foreground(m.palette.Dim).Background(m.palette.Bg).Render("Enter: submit · Esc: cancel")
	body += bg.Width(innerW).Render(" ") + "\n" + bg.Width(innerW).Render(help)
	return box.Render(body)
}

// Title implements Modal.
func (m TextInputModel) Title() string { return m.prompt }
