package modals

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/ui/theme"
)

// RunSubmittedMsg is emitted by RunFormModel when the user submits.
type RunSubmittedMsg struct {
	Opts cli.RunOpts
}

// RunCancelledMsg is emitted when the user cancels the run form.
type RunCancelledMsg struct{}

// RunFormModel is a multi-field form for the :run modal.
//
// Fields (in order): name, image, ports, env, volumes, interactive,
// tty, detach. Tab cycles between fields; Ctrl+Enter submits. Enter
// also submits when on the last toggle. Esc cancels.
type RunFormModel struct {
	name        textinput.Model
	image       textinput.Model
	ports       textinput.Model
	env         textinput.Model
	volumes     textinput.Model
	interactive bool
	tty         bool
	detach      bool
	focus       int
	palette     theme.Palette
}

const (
	runFieldName        = 0
	runFieldImage       = 1
	runFieldPorts       = 2
	runFieldEnv         = 3
	runFieldVolumes     = 4
	runFieldInteractive = 5
	runFieldTTY         = 6
	runFieldDetach      = 7
	runFieldCount       = 8
)

// NewRunForm creates a new run form modal pre-filled with the given image
// reference (may be empty).
func NewRunForm(imageHint string, p theme.Palette) RunFormModel {
	mk := func(prompt, placeholder string) textinput.Model {
		t := textinput.New()
		t.Prompt = prompt
		t.Placeholder = placeholder
		t.CharLimit = 256
		t.SetWidth(50)
		styleTextInput(&t, p)
		return t
	}
	name := mk("Name:    ", "auto")
	image := mk("Image:   ", "ghcr.io/me/api:latest")
	ports := mk("Ports:   ", "8080:8080,443:443")
	env := mk("Env:     ", "KEY=val,OTHER=2")
	volumes := mk("Volumes: ", "src:dst,data:/data")
	if imageHint != "" {
		image.SetValue(imageHint)
	}
	m := RunFormModel{
		name:    name,
		image:   image,
		ports:   ports,
		env:     env,
		volumes: volumes,
		detach:  true,
		palette: p,
	}
	if imageHint != "" {
		m.focus = runFieldName
	}
	m.applyFocus()
	return m
}

// Init implements Modal.
func (m RunFormModel) Init() tea.Cmd { return textinput.Blink }

// Update implements Modal.
func (m RunFormModel) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return m, tea.Batch(
				func() tea.Msg { return RunCancelledMsg{} },
				CloseModal(),
			)
		case "tab":
			m.focus = (m.focus + 1) % runFieldCount
			m.applyFocus()
			return m, nil
		case "shift+tab":
			m.focus = (m.focus + runFieldCount - 1) % runFieldCount
			m.applyFocus()
			return m, nil
		case "ctrl+d":
			// Convenient shortcut: Ctrl-D submits.
			return m.submit()
		}
		switch msg.String() {
		case "ctrl+s", "ctrl+enter":
			return m.submit()
		case "space":
			// Toggle on bool fields
			switch m.focus {
			case runFieldInteractive:
				m.interactive = !m.interactive
				return m, nil
			case runFieldTTY:
				m.tty = !m.tty
				return m, nil
			case runFieldDetach:
				m.detach = !m.detach
				return m, nil
			}
		case "enter":
			// Enter advances except on the last field, where it submits
			if m.focus == runFieldCount-1 {
				return m.submit()
			}
			m.focus = (m.focus + 1) % runFieldCount
			m.applyFocus()
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.focus {
	case runFieldName:
		m.name, cmd = m.name.Update(msg)
	case runFieldImage:
		m.image, cmd = m.image.Update(msg)
	case runFieldPorts:
		m.ports, cmd = m.ports.Update(msg)
	case runFieldEnv:
		m.env, cmd = m.env.Update(msg)
	case runFieldVolumes:
		m.volumes, cmd = m.volumes.Update(msg)
	}
	return m, cmd
}

func (m *RunFormModel) applyFocus() {
	m.name.Blur()
	m.image.Blur()
	m.ports.Blur()
	m.env.Blur()
	m.volumes.Blur()
	switch m.focus {
	case runFieldName:
		m.name.Focus()
	case runFieldImage:
		m.image.Focus()
	case runFieldPorts:
		m.ports.Focus()
	case runFieldEnv:
		m.env.Focus()
	case runFieldVolumes:
		m.volumes.Focus()
	}
}

func (m RunFormModel) submit() (Modal, tea.Cmd) {
	opts := cli.RunOpts{
		Name:        strings.TrimSpace(m.name.Value()),
		Image:       strings.TrimSpace(m.image.Value()),
		Ports:       splitCSV(m.ports.Value()),
		Env:         splitCSV(m.env.Value()),
		Volumes:     splitCSV(m.volumes.Value()),
		Interactive: m.interactive,
		TTY:         m.tty,
		Detach:      m.detach,
	}
	return m, tea.Batch(
		func() tea.Msg { return RunSubmittedMsg{Opts: opts} },
		CloseModal(),
	)
}

// View implements Modal.
func (m RunFormModel) View(width, height int) string {
	const boxW = 72
	innerW := boxW - 6 // border (2) + padding (4)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Border).
		BorderBackground(m.palette.Bg).
		Background(m.palette.Bg).
		Foreground(m.palette.Fg).
		Padding(1, 2).
		Width(boxW)

	bg := lipgloss.NewStyle().Foreground(m.palette.Fg).Background(m.palette.Bg)
	title := lipgloss.NewStyle().Bold(true).Foreground(m.palette.HeaderFg).Background(m.palette.Bg).Render("Run container")
	help := lipgloss.NewStyle().Foreground(m.palette.Dim).Background(m.palette.Bg).Render(
		"Tab cycle · Space toggle · Ctrl-Enter submit · Esc cancel",
	)

	body := bg.Width(innerW).Render(title) + "\n" +
		bg.Width(innerW).Render(" ") + "\n" +
		renderTextInput(m.name, m.palette, innerW) + "\n" +
		renderTextInput(m.image, m.palette, innerW) + "\n" +
		renderTextInput(m.ports, m.palette, innerW) + "\n" +
		renderTextInput(m.env, m.palette, innerW) + "\n" +
		renderTextInput(m.volumes, m.palette, innerW) + "\n" +
		bg.Width(innerW).Render(m.checkbox(runFieldInteractive, "Interactive", m.interactive)) + "\n" +
		bg.Width(innerW).Render(m.checkbox(runFieldTTY, "TTY", m.tty)) + "\n" +
		bg.Width(innerW).Render(m.checkbox(runFieldDetach, "Detach", m.detach)) + "\n" +
		bg.Width(innerW).Render(" ") + "\n" +
		bg.Width(innerW).Render(help)

	return box.Render(body)
}

func (m RunFormModel) checkbox(field int, label string, on bool) string {
	mark := "[ ]"
	if on {
		mark = "[x]"
	}
	cursor := "  "
	if m.focus == field {
		cursor = "→ "
	}
	return cursor + mark + " " + label
}

// Title implements Modal.
func (m RunFormModel) Title() string { return "Run container" }

// splitCSV splits a comma-separated string and trims whitespace, dropping
// empty entries. Used by the run/build form modals to parse repeatable
// flags entered on a single line.
func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
