package skinx

import (
	"testing"

	"github.com/torosent/c9s/internal/ui/theme"
)

func TestTableStyles_UsesPalette(t *testing.T) {
	p := theme.DefaultDark()
	s := TableStyles(p)
	// Header should use Fg/Bg from palette, not bubbles defaults.
	if got := s.Header.GetForeground(); got != p.Fg {
		t.Errorf("Header.Foreground = %v, want %v", got, p.Fg)
	}
	if got := s.Header.GetBackground(); got != p.Bg {
		t.Errorf("Header.Background = %v, want %v", got, p.Bg)
	}
	// Cell intentionally has no Fg/Bg — see TableStyles comment. The body
	// wrapper paints palette.Bg/Fg; leaving cells unstyled lets the
	// Selected.Render bg actually show through on the cursor row.
	if rendered := s.Cell.Render("X"); rendered != "X" {
		t.Errorf("Cell should be unstyled (render plain), got %q", rendered)
	}
	// Selected uses Accent for high contrast regardless of skin's cursor
	// color (which can be near-bg in some imported themes).
	if got := s.Selected.GetForeground(); got != p.Bg {
		t.Errorf("Selected.Foreground = %v, want palette.Bg %v", got, p.Bg)
	}
	if got := s.Selected.GetBackground(); got != p.Accent {
		t.Errorf("Selected.Background = %v, want palette.Accent %v", got, p.Accent)
	}
}

func TestTableStyles_HeaderBoldAndBordered(t *testing.T) {
	p := theme.DefaultDark()
	s := TableStyles(p)
	if !s.Header.GetBold() {
		t.Error("Header should be bold so it stays readable when bg matches body")
	}
	if !s.Header.GetBorderBottom() {
		t.Error("Header should have a bottom border to separate from rows")
	}
	if got := s.Header.GetBorderBottomBackground(); got != p.Bg {
		t.Errorf("Header border bg = %v, want palette.Bg %v", got, p.Bg)
	}
}
