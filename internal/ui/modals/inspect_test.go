package modals

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestInspectPrettyPrintsValidJSON(t *testing.T) {
	p := theme.DefaultDark()
	jsonBytes := []byte(`{"a":1,"b":"test"}`)
	m := NewInspect("Container Details", jsonBytes, p)

	v := m.View(80, 20)
	stripped := stripANSI(v)

	// Should contain pretty-printed JSON with indentation
	if !strings.Contains(stripped, `"a": 1`) {
		t.Errorf("expected pretty-printed JSON with '\"a\": 1', got: %s", stripped)
	}
	if !strings.Contains(stripped, `"b": "test"`) {
		t.Errorf("expected '\"b\": \"test\"' in view")
	}
}

func TestInspectRawOnInvalid(t *testing.T) {
	p := theme.DefaultDark()
	jsonBytes := []byte(`not json at all`)
	m := NewInspect("Invalid JSON", jsonBytes, p)

	v := m.View(80, 20)
	stripped := stripANSI(v)

	// Should contain the raw input when JSON parsing fails
	if !strings.Contains(stripped, "not json at all") {
		t.Errorf("expected raw input 'not json at all' in view, got: %s", stripped)
	}
}

func TestInspectEscCloses(t *testing.T) {
	p := theme.DefaultDark()
	m := NewInspect("Test", []byte(`{"a":1}`), p)

	keyMsg := tea.KeyPressMsg{Code: tea.KeyEsc}
	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Fatal("expected Update to return a cmd on Esc keypress")
	}

	msg := cmd()
	if _, ok := msg.(CloseModalMsg); !ok {
		t.Errorf("expected CloseModalMsg, got %T", msg)
	}
}

func TestInspectQCloses(t *testing.T) {
	p := theme.DefaultDark()
	m := NewInspect("Test", []byte(`{"a":1}`), p)

	keyMsg := tea.KeyPressMsg{Code: 'q', Text: "q"}
	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Fatal("expected Update to return a cmd on 'q' keypress")
	}

	msg := cmd()
	if _, ok := msg.(CloseModalMsg); !ok {
		t.Errorf("expected CloseModalMsg, got %T", msg)
	}
}
