package modals

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/ui/theme"
)

// BuildSubmittedMsg is emitted when the user submits the build form.
type BuildSubmittedMsg struct {
	Opts cli.BuildOpts
}

// BuildCancelledMsg is emitted when the user cancels the build form.
type BuildCancelledMsg struct{}

// BuildFormModel is a multi-field form for the :build modal.
//
// Fields: path (context dir), tag (-t), containerfile (-f), platform.
// Tab cycles fields; Ctrl-Enter / Ctrl-D submits; Esc cancels.
type BuildFormModel struct {
	path          textinput.Model
	tag           textinput.Model
	containerfile textinput.Model
	platform      textinput.Model
	focus         int
	palette       theme.Palette
}

const (
	buildFieldPath          = 0
	buildFieldTag           = 1
	buildFieldContainerfile = 2
	buildFieldPlatform      = 3
	buildFieldCount         = 4
)

// NewBuildForm creates a new build form modal pre-filled with the given
// path (may be empty to start at the path field).
func NewBuildForm(pathHint string, p theme.Palette) BuildFormModel {
	mk := func(prompt, placeholder string) textinput.Model {
		t := textinput.New()
		t.Prompt = prompt
		t.Placeholder = placeholder
		t.CharLimit = 256
		t.SetWidth(50)
		styleTextInput(&t, p)
		return t
	}
	path := mk("Path:          ", "./")
	tag := mk("Tag (-t):      ", "ghcr.io/me/api:latest")
	cf := mk("Containerfile: ", "Containerfile")
	plat := mk("Platform:      ", "linux/arm64")
	if pathHint != "" {
		path.SetValue(pathHint)
	}
	m := BuildFormModel{
		path:          path,
		tag:           tag,
		containerfile: cf,
		platform:      plat,
		palette:       p,
	}
	if pathHint != "" {
		m.focus = buildFieldTag
	}
	m.applyFocus()
	return m
}

// Init implements Modal.
func (m BuildFormModel) Init() tea.Cmd { return textinput.Blink }

// Update implements Modal.
func (m BuildFormModel) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return m, tea.Batch(
				func() tea.Msg { return BuildCancelledMsg{} },
				CloseModal(),
			)
		case "tab":
			m.focus = (m.focus + 1) % buildFieldCount
			m.applyFocus()
			return m, nil
		case "shift+tab":
			m.focus = (m.focus + buildFieldCount - 1) % buildFieldCount
			m.applyFocus()
			return m, nil
		case "ctrl+d", "ctrl+s":
			return m.submit()
		}
		switch msg.String() {
		case "ctrl+enter":
			return m.submit()
		case "enter":
			if m.focus == buildFieldCount-1 {
				return m.submit()
			}
			m.focus = (m.focus + 1) % buildFieldCount
			m.applyFocus()
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.focus {
	case buildFieldPath:
		m.path, cmd = m.path.Update(msg)
	case buildFieldTag:
		m.tag, cmd = m.tag.Update(msg)
	case buildFieldContainerfile:
		m.containerfile, cmd = m.containerfile.Update(msg)
	case buildFieldPlatform:
		m.platform, cmd = m.platform.Update(msg)
	}
	return m, cmd
}

func (m *BuildFormModel) applyFocus() {
	m.path.Blur()
	m.tag.Blur()
	m.containerfile.Blur()
	m.platform.Blur()
	switch m.focus {
	case buildFieldPath:
		m.path.Focus()
	case buildFieldTag:
		m.tag.Focus()
	case buildFieldContainerfile:
		m.containerfile.Focus()
	case buildFieldPlatform:
		m.platform.Focus()
	}
}

func (m BuildFormModel) submit() (Modal, tea.Cmd) {
	opts := cli.BuildOpts{
		ContextPath:       strings.TrimSpace(m.path.Value()),
		Tag:               strings.TrimSpace(m.tag.Value()),
		ContainerfilePath: strings.TrimSpace(m.containerfile.Value()),
		Platform:          strings.TrimSpace(m.platform.Value()),
	}
	return m, tea.Batch(
		func() tea.Msg { return BuildSubmittedMsg{Opts: opts} },
		CloseModal(),
	)
}

// View implements Modal.
func (m BuildFormModel) View(width, height int) string {
	const boxW = 72
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
	title := lipgloss.NewStyle().Bold(true).Foreground(m.palette.HeaderFg).Background(m.palette.Bg).Render("Build image")
	help := lipgloss.NewStyle().Foreground(m.palette.Dim).Background(m.palette.Bg).Render(
		"Tab cycle · Ctrl-Enter submit · Esc cancel",
	)

	body := bg.Width(innerW).Render(title) + "\n" +
		bg.Width(innerW).Render(" ") + "\n" +
		renderTextInput(m.path, m.palette, innerW) + "\n" +
		renderTextInput(m.tag, m.palette, innerW) + "\n" +
		renderTextInput(m.containerfile, m.palette, innerW) + "\n" +
		renderTextInput(m.platform, m.palette, innerW) + "\n" +
		bg.Width(innerW).Render(" ") + "\n" +
		bg.Width(innerW).Render(help)

	return box.Render(body)
}

// Title implements Modal.
func (m BuildFormModel) Title() string { return "Build image" }
