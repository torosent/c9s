// Package skinx provides table-style helpers that respect the c9s skin palette.
package skinx

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"github.com/torosent/c9s/internal/ui/theme"
)

// TableStyles returns a bubbles/table.Styles tree that paints every cell with
// the given palette: header, cell, and selected row all match the skin so a
// light theme renders consistently end-to-end.
//
// Both Header and Cell explicitly clear Padding. Bubbles' DefaultStyles
// applies Padding(0, 1) to both — but headersView() and renderRow() each
// pre-render their content to col.Width via lipgloss.NewStyle().Width().
// When the outer Header style adds Padding(0, 1), the header cell becomes
// col.Width + 2 wide while the data cell stays at col.Width, causing a
// 14-col mismatch over the 7-column containers table that wraps the
// header onto two lines and shifts header labels relative to data
// columns. Stripping Padding on both sides keeps them aligned; we add
// inter-column visual separation by prefixing every column title and
// every row value with a single space at the call site.
func TableStyles(p theme.Palette) table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(p.Border).
		BorderBackground(p.Bg).
		BorderBottom(true).
		Foreground(p.Fg).
		Background(p.Bg).
		Padding(0, 0).
		Bold(true)
	// Cell style intentionally has NO Foreground or Background (and no
	// Padding). bubbles/table wraps cell content with the cell style THEN
	// the selected-row style. If we set Cell.Background, the per-cell bg
	// "wins" against Selected.Render's outer bg — making the cursor row's
	// selection invisible. Letting cells be unstyled means the body
	// wrapper's palette.Bg/Fg shows through non-selected rows, and
	// Selected.Render's Background(Accent) correctly paints the entire
	// cursor row.
	s.Cell = lipgloss.NewStyle()
	s.Selected = lipgloss.NewStyle().
		Foreground(p.Bg).
		Background(p.Accent).
		Bold(true)
	return s
}

// BorderedBox wraps content in a single-line border with a centered k9s-style
// title `Title(filter)[count]`. Width and height are the outer dimensions.
func BorderedBox(p theme.Palette, title, filter string, count, width, height int, content string) string {
	if width <= 0 || height <= 0 {
		return content
	}

	titleStyle := lipgloss.NewStyle().Foreground(p.Accent).Bold(true).Background(p.Bg)
	parenStyle := lipgloss.NewStyle().Foreground(p.Fg).Background(p.Bg)
	countStyle := lipgloss.NewStyle().Foreground(p.Accent).Bold(true).Background(p.Bg)

	if filter == "" {
		filter = "all"
	}
	header := titleStyle.Render(title) +
		parenStyle.Render("(") + countStyle.Render(filter) + parenStyle.Render(")") +
		parenStyle.Render("[") + countStyle.Render(fmt.Sprintf("%d", count)) + parenStyle.Render("]")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Accent).
		BorderBackground(p.Bg).
		Background(p.Bg).
		Foreground(p.Fg).
		Width(width-2).
		Height(height-2).
		Padding(0, 1)

	// Render content first, then overlay the title onto the top border.
	rendered := box.Render(content)
	// Replace the first border line's center with the header text.
	return overlayTitle(rendered, header, p.Accent, p.Bg)
}

// overlayTitle splices the styled title into the top-border row.
func overlayTitle(boxed, header string, border, bg lipgloss.Color) string {
	// Find first newline (top border row).
	for i := 0; i < len(boxed); i++ {
		if boxed[i] == '\n' {
			top := boxed[:i]
			rest := boxed[i:]
			// Inject title centered into the border line.
			// Strip ANSI to measure visible width:
			vis := lipgloss.Width(top)
			titleVis := lipgloss.Width(header)
			leftPad := (vis - titleVis) / 2
			if leftPad < 4 {
				leftPad = 4
			}
			borderStyle := lipgloss.NewStyle().Foreground(border).Background(bg)
			// Use box drawing chars: corners + horizontal
			leftBorder := borderStyle.Render("╭" + repeat("─", leftPad-1))
			rightLen := vis - leftPad - titleVis - 1
			if rightLen < 1 {
				rightLen = 1
			}
			rightBorder := borderStyle.Render(repeat("─", rightLen) + "╮")
			return leftBorder + header + rightBorder + rest
		}
	}
	return boxed
}

func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
