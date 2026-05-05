package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/torosent/c9s/internal/ui/theme"
)

// splashLogo is the same ANSI Shadow figlet 'c9s' as the header banner
// — keeps the brand consistent between the splash screen and the main view.
const splashLogo = ` ██████╗ ██████╗ ███████╗
██╔════╝██╔═══██╗██╔════╝
██║     ╚██████╔╝███████╗
██║      ╚═══██║ ╚════██║
╚██████╗  █████╔╝███████║
 ╚═════╝  ╚════╝ ╚══════╝`

// SplashDoneMsg is emitted when the splash should be dismissed (any keypress).
type SplashDoneMsg struct{}

// SplashModel is the startup splash screen.
type SplashModel struct {
	palette theme.Palette
	version string
	tip     string
	width   int
	height  int
}

// NewSplash constructs a SplashModel with a tip.
func NewSplash(p theme.Palette, version string) SplashModel {
	tips := []string{
		"Press : to open command palette",
		"Use ? for help on any screen",
		"Press b to bookmark resources",
		":errors shows error log history",
		":xray displays resource relationships",
		":pulses shows live system dashboard",
		":pinned shows bookmarked resources",
	}
	// Simple day-based rotation
	dayOfYear := time.Now().YearDay()
	tip := tips[dayOfYear%len(tips)]
	return SplashModel{palette: p, version: version, tip: tip}
}

// Init returns nil — the splash has nothing to do until input arrives.
func (m SplashModel) Init() tea.Cmd { return nil }

// Update handles keypresses and window resize.
func (m SplashModel) Update(msg tea.Msg) (SplashModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		_ = msg
		return m, func() tea.Msg { return SplashDoneMsg{} }
	}
	return m, nil
}

// View renders the splash centered in the available space.
func (m SplashModel) View() string {
	p := m.palette
	bg := lipgloss.NewStyle().Background(p.Bg).Foreground(p.Fg)
	logoStyle := lipgloss.NewStyle().Foreground(p.Accent).Background(p.Bg).Bold(true)

	// Render each logo line individually with bg + fixed width — same
	// pattern the banner uses for the right-corner logo. JoinVertical
	// of Width()-rendered cells produces aligned output without the
	// per-cell pad-with-bare-space artifacts.
	logoLines := strings.Split(strings.Trim(splashLogo, "\n"), "\n")
	maxW := 0
	for _, line := range logoLines {
		if w := lipgloss.Width(line); w > maxW {
			maxW = w
		}
	}
	logoRows := make([]string, len(logoLines))
	for i, line := range logoLines {
		logoRows[i] = bg.Width(maxW).Render(logoStyle.Render(line))
	}
	logo := lipgloss.JoinVertical(lipgloss.Left, logoRows...)

	tagline := bg.Bold(true).Render("Apple Containers TUI")
	sub := bg.Render(m.version)
	tipLine := bg.Render("Tip: " + m.tip)
	hint := lipgloss.NewStyle().Foreground(p.Accent).Background(p.Bg).Bold(true).Render("press any key to continue")

	// Pad every body line to the same width with bg-styled chars so
	// JoinVertical(Center, ...) doesn't insert unstyled centring spaces.
	// For multi-line items (the logo), split first so each individual
	// line gets centred independently.
	bodyLines := []string{}
	for _, item := range []string{logo, tagline, "", sub, "", tipLine, "", hint} {
		bodyLines = append(bodyLines, strings.Split(item, "\n")...)
	}
	bodyW := 0
	for _, l := range bodyLines {
		if w := lipgloss.Width(l); w > bodyW {
			bodyW = w
		}
	}
	bgPad := lipgloss.NewStyle().Background(p.Bg)
	for i, l := range bodyLines {
		if l == "" {
			bodyLines[i] = bgPad.Width(bodyW).Render(" ")
			continue
		}
		w := lipgloss.Width(l)
		if w < bodyW {
			leftPad := (bodyW - w) / 2
			rightPad := bodyW - w - leftPad
			bodyLines[i] = bgPad.Render(strings.Repeat(" ", leftPad)) + l + bgPad.Render(strings.Repeat(" ", rightPad))
		}
	}
	body := strings.Join(bodyLines, "\n")
	if m.width == 0 || m.height == 0 {
		return body
	}
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		body,
		lipgloss.WithWhitespaceBackground(p.Bg),
		lipgloss.WithWhitespaceForeground(p.Bg),
	)
}
