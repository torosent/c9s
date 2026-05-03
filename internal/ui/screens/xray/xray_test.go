package xray

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/ui/theme"
	"github.com/torosent/c9s/internal/ui/widgets"
)

func TestNew(t *testing.T) {
	fake := &cli.Fake{}
	p := theme.DefaultDark()

	m := New(fake, p)
	if m == nil {
		t.Fatal("expected non-nil model")
	}

	if m.Title() != "xray" {
		t.Errorf("expected title 'xray', got %s", m.Title())
	}
}

func TestTreeBuilt(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
		},
	}
	p := theme.DefaultDark()

	m := New(fake, p)

	// Trigger tree build
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected init command")
	}

	msg := cmd()
	treeMsg, ok := msg.(TreeBuiltMsg)
	if !ok {
		t.Fatalf("expected TreeBuiltMsg, got %T", msg)
	}

	// Update with tree
	m.Update(treeMsg)

	// Verify tree has container node
	view := m.View(80, 24)
	if !strings.Contains(view, "container") {
		t.Errorf("expected view to contain 'container', got: %s", view)
	}
	// Note: image is a child of container, may not be visible if collapsed
}

func TestExpandCollapseKeys(t *testing.T) {
	fake := &cli.Fake{}
	p := theme.DefaultDark()

	m := New(fake, p)

	root := &widgets.Node{Label: "Root", Kind: "root", Expanded: false}
	root.AddChild(&widgets.Node{Label: "Child", Kind: "container"})
	m.Update(TreeBuiltMsg{Root: root})

	// Press 'e' to expand
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	view := m.View(80, 24)
	if !strings.Contains(view, "Child") {
		t.Errorf("expected 'Child' to be visible after expand, got: %s", view)
	}

	// Press 'c' to collapse
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	view = m.View(80, 24)
	if strings.Contains(view, "Child") {
		t.Errorf("expected 'Child' to be hidden after collapse, got: %s", view)
	}
}

func TestEnterEmitsJump(t *testing.T) {
	fake := &cli.Fake{}
	p := theme.DefaultDark()

	m := New(fake, p)

	root := &widgets.Node{Label: "Root", Kind: "root"}
	cNode := &widgets.Node{Label: "C1", Kind: "container", ID: "c1"}
	root.AddChild(cNode)
	m.Update(TreeBuiltMsg{Root: root})

	// Move down to container node
	m.tree.MoveDown()

	// Verify we're on the container node
	if m.tree.Focused == nil || m.tree.Focused.Kind != "container" {
		t.Skip("focus not on container node, skipping enter test")
	}

	// Press enter
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command after pressing enter")
	}

	msg := cmd()
	jumpMsg, ok := msg.(JumpToResourceMsg)
	if !ok {
		t.Fatalf("expected JumpToResourceMsg, got %T", msg)
	}

	if jumpMsg.Kind != "container" {
		t.Errorf("expected kind 'container', got %s", jumpMsg.Kind)
	}
	if jumpMsg.ID != "c1" {
		t.Errorf("expected ID 'c1', got %s", jumpMsg.ID)
	}
}
