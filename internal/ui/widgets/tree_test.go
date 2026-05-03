package widgets

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewTree(t *testing.T) {
	root := &Node{Label: "Root", Kind: "root"}
	tree := NewTree(root)

	if tree.Root != root {
		t.Error("expected tree root to be set")
	}
	if tree.Focused == nil {
		t.Error("expected focused to be set to root")
	}
}

func TestMoveUpDown(t *testing.T) {
	root := &Node{Label: "Root", Expanded: true}
	child1 := &Node{Label: "Child1"}
	child2 := &Node{Label: "Child2"}
	root.AddChild(child1)
	root.AddChild(child2)

	tree := NewTree(root)

	// Initially focused on root
	if tree.Focused != root {
		t.Errorf("expected focused on root, got %v", tree.Focused)
	}

	// Move down
	tree.MoveDown()
	if tree.Focused != child1 {
		t.Errorf("expected focused on child1, got %v", tree.Focused)
	}

	// Move down again
	tree.MoveDown()
	if tree.Focused != child2 {
		t.Errorf("expected focused on child2, got %v", tree.Focused)
	}

	// Move up
	tree.MoveUp()
	if tree.Focused != child1 {
		t.Errorf("expected focused on child1 after move up, got %v", tree.Focused)
	}
}

func TestExpandCollapse(t *testing.T) {
	root := &Node{Label: "Root", Expanded: false}
	child := &Node{Label: "Child"}
	root.AddChild(child)

	tree := NewTree(root)

	// Initially collapsed
	if len(tree.flat) != 1 {
		t.Errorf("expected 1 visible node (root), got %d", len(tree.flat))
	}

	// Expand
	tree.Expand()
	if !root.Expanded {
		t.Error("expected root to be expanded")
	}
	if len(tree.flat) != 2 {
		t.Errorf("expected 2 visible nodes after expand, got %d", len(tree.flat))
	}

	// Collapse
	tree.Collapse()
	if root.Expanded {
		t.Error("expected root to be collapsed")
	}
	if len(tree.flat) != 1 {
		t.Errorf("expected 1 visible node after collapse, got %d", len(tree.flat))
	}
}

func TestView(t *testing.T) {
	root := &Node{Label: "Root", Expanded: true}
	child := &Node{Label: "Child"}
	root.AddChild(child)

	tree := NewTree(root)

	view := tree.View()
	if !strings.Contains(view, "Root") {
		t.Errorf("expected view to contain 'Root', got: %s", view)
	}
	if !strings.Contains(view, "Child") {
		t.Errorf("expected view to contain 'Child', got: %s", view)
	}
	if !strings.Contains(view, ">") {
		t.Errorf("expected view to contain focus marker '>', got: %s", view)
	}
}

func TestUpdate(t *testing.T) {
	root := &Node{Label: "Root", Expanded: true}
	child := &Node{Label: "Child"}
	root.AddChild(child)

	tree := NewTree(root)

	// Test key handling
	tree, _ = tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if tree.Focused != child {
		t.Errorf("expected focus to move to child after 'j', got %v", tree.Focused)
	}

	tree, _ = tree.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if tree.Focused != root {
		t.Errorf("expected focus to move to root after 'k', got %v", tree.Focused)
	}
}

func TestExpandCollapseAll(t *testing.T) {
	root := &Node{Label: "Root", Expanded: false}
	child1 := &Node{Label: "Child1", Expanded: false}
	child2 := &Node{Label: "Child2"}
	root.AddChild(child1)
	child1.AddChild(child2)

	tree := NewTree(root)

	// Expand all
	tree.ExpandAll()
	if !root.Expanded || !child1.Expanded {
		t.Error("expected all nodes to be expanded")
	}

	// Collapse all
	tree.CollapseAll()
	if root.Expanded || child1.Expanded {
		t.Error("expected all nodes to be collapsed")
	}
}
