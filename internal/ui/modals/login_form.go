package modals

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/torosent/c9s/internal/ui/theme"
)

// LoginRequest carries the credentials the user entered.
type LoginRequest struct {
	Host     string
	Username string
	Password string
}

// LoginResultMsg is emitted when the user submits the login modal.
type LoginResultMsg struct {
	Result LoginRequest
}

// LoginCancelledMsg is emitted when the user dismisses the modal with Esc.
type LoginCancelledMsg struct{}

// LoginModel is a 3-field form for `container registry login`.
// The password input is rendered with `*` in place of the typed runes;
// the underlying value is held in textinput.Value().
type LoginModel struct {
	host     textinput.Model
	user     textinput.Model
	password textinput.Model
	focus    int
	palette  theme.Palette
}

// NewLogin creates a new login modal. If hostHint is non-empty, it pre-fills
// the host field. Focus starts on the first empty field.
func NewLogin(hostHint string, p theme.Palette) LoginModel {
	host := textinput.New()
	host.Placeholder = "ghcr.io"
	host.Prompt = "Host:     "
	host.CharLimit = 128
	host.Width = 40
	styleTextInput(&host, p)
	if hostHint != "" {
		host.SetValue(hostHint)
	}

	user := textinput.New()
	user.Placeholder = "username"
	user.Prompt = "User:     "
	user.CharLimit = 128
	user.Width = 40
	styleTextInput(&user, p)

	pass := textinput.New()
	pass.Placeholder = "(typed characters are masked)"
	pass.Prompt = "Password: "
	pass.CharLimit = 256
	pass.Width = 40
	pass.EchoMode = textinput.EchoPassword
	pass.EchoCharacter = '*'
	styleTextInput(&pass, p)

	m := LoginModel{
		host:     host,
		user:     user,
		password: pass,
		palette:  p,
	}
	if hostHint != "" {
		m.focus = 1
	}
	m.applyFocus()
	return m
}

// Init implements Modal.
func (m LoginModel) Init() tea.Cmd { return textinput.Blink }

// Update implements Modal.
func (m LoginModel) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return m, tea.Batch(
				func() tea.Msg { return LoginCancelledMsg{} },
				CloseModal(),
			)
		case tea.KeyEnter:
			if m.focus < 2 {
				m.focus++
				m.applyFocus()
				return m, nil
			}
			return m, tea.Batch(
				func() tea.Msg {
					return LoginResultMsg{Result: LoginRequest{
						Host:     strings.TrimSpace(m.host.Value()),
						Username: strings.TrimSpace(m.user.Value()),
						Password: m.password.Value(),
					}}
				},
				CloseModal(),
			)
		case tea.KeyTab:
			m.focus = (m.focus + 1) % 3
			m.applyFocus()
			return m, nil
		case tea.KeyShiftTab:
			m.focus = (m.focus + 2) % 3
			m.applyFocus()
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.focus {
	case 0:
		m.host, cmd = m.host.Update(msg)
	case 1:
		m.user, cmd = m.user.Update(msg)
	case 2:
		m.password, cmd = m.password.Update(msg)
	}
	return m, cmd
}

func (m *LoginModel) applyFocus() {
	m.host.Blur()
	m.user.Blur()
	m.password.Blur()
	switch m.focus {
	case 0:
		m.host.Focus()
	case 1:
		m.user.Focus()
	case 2:
		m.password.Focus()
	}
}

// View implements Modal.
func (m LoginModel) View(width, height int) string {
	const boxW = 60
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
	title := lipgloss.NewStyle().Bold(true).Foreground(m.palette.HeaderFg).Background(m.palette.Bg).Render("Registry login")
	help := lipgloss.NewStyle().Foreground(m.palette.Dim).Background(m.palette.Bg).Render(
		"Tab/Shift-Tab cycle fields · Enter advances/submits · Esc cancel",
	)

	body := bg.Width(innerW).Render(title) + "\n" +
		bg.Width(innerW).Render(" ") + "\n" +
		renderTextInput(m.host, m.palette, innerW) + "\n" +
		renderTextInput(m.user, m.palette, innerW) + "\n" +
		renderTextInput(m.password, m.palette, innerW) + "\n" +
		bg.Width(innerW).Render(" ") + "\n" +
		bg.Width(innerW).Render(help)
	return box.Render(body)
}

// Title implements Modal.
func (m LoginModel) Title() string { return "Registry login" }
