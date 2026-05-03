package breadcrumbs

import (
	"testing"
)

func TestNewTrail(t *testing.T) {
	trail := New()
	if trail.Len() != 0 {
		t.Errorf("expected new trail to be empty, got len=%d", trail.Len())
	}
}

func TestPushPop(t *testing.T) {
	trail := New()
	c1 := Crumb{Label: "containers", Screen: "containers"}
	c2 := Crumb{Label: "logs", Screen: "logs"}

	trail.Push(c1)
	if trail.Len() != 1 {
		t.Errorf("expected len=1 after push, got %d", trail.Len())
	}

	trail.Push(c2)
	if trail.Len() != 2 {
		t.Errorf("expected len=2 after second push, got %d", trail.Len())
	}

	// Pop should return c2
	popped, ok := trail.Pop()
	if !ok {
		t.Fatal("expected Pop to succeed")
	}
	if popped.Label != "logs" {
		t.Errorf("expected to pop 'logs', got %s", popped.Label)
	}
	if trail.Len() != 1 {
		t.Errorf("expected len=1 after pop, got %d", trail.Len())
	}

	// Pop should return c1
	popped, ok = trail.Pop()
	if !ok {
		t.Fatal("expected second Pop to succeed")
	}
	if popped.Label != "containers" {
		t.Errorf("expected to pop 'containers', got %s", popped.Label)
	}
	if trail.Len() != 0 {
		t.Errorf("expected len=0 after second pop, got %d", trail.Len())
	}

	// Pop on empty should return false
	_, ok = trail.Pop()
	if ok {
		t.Error("expected Pop on empty trail to return false")
	}
}

func TestTop(t *testing.T) {
	trail := New()
	c1 := Crumb{Label: "containers", Screen: "containers"}
	c2 := Crumb{Label: "logs", Screen: "logs"}

	// Top on empty
	_, ok := trail.Top()
	if ok {
		t.Error("expected Top on empty trail to return false")
	}

	trail.Push(c1)
	trail.Push(c2)

	// Top should return c2 without removing it
	top, ok := trail.Top()
	if !ok {
		t.Fatal("expected Top to succeed")
	}
	if top.Label != "logs" {
		t.Errorf("expected top to be 'logs', got %s", top.Label)
	}
	if trail.Len() != 2 {
		t.Errorf("expected len=2 after Top, got %d", trail.Len())
	}
}

func TestRenderShort(t *testing.T) {
	trail := New()
	trail.Push(Crumb{Label: "containers", Screen: "containers"})
	trail.Push(Crumb{Label: "logs", Screen: "logs"})

	rendered := trail.Render(100)
	expected := "containers > logs"
	if rendered != expected {
		t.Errorf("expected %q, got %q", expected, rendered)
	}
}

func TestRenderTruncated(t *testing.T) {
	trail := New()
	trail.Push(Crumb{Label: "containers", Screen: "containers"})
	trail.Push(Crumb{Label: "images", Screen: "images"})
	trail.Push(Crumb{Label: "volumes", Screen: "volumes"})
	trail.Push(Crumb{Label: "networks", Screen: "networks"})

	// Should truncate middle
	rendered := trail.Render(30)
	// Should contain first and last with ellipsis
	if !(contains(rendered, "containers") && contains(rendered, "networks") && contains(rendered, "…")) {
		t.Errorf("expected truncated trail with ellipsis, got %q", rendered)
	}
}

func TestRenderEmpty(t *testing.T) {
	trail := New()
	rendered := trail.Render(100)
	if rendered != "" {
		t.Errorf("expected empty string for empty trail, got %q", rendered)
	}
}

func TestRenderSingleCrumb(t *testing.T) {
	trail := New()
	trail.Push(Crumb{Label: "containers", Screen: "containers"})

	rendered := trail.Render(100)
	if rendered != "containers" {
		t.Errorf("expected 'containers', got %q", rendered)
	}

	// Truncate single crumb - check display width, not byte length
	rendered = trail.Render(5)
	// The ellipsis may be multibyte, so check that it's visually <=5 wide
	// We can't easily test lipgloss.Width here, so just verify it's truncated
	if !contains(rendered, "…") && len(rendered) > 5 {
		t.Errorf("expected truncated string, got %q", rendered)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && anySubstring(s, substr))
}

func anySubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
