// Package widgets provides reusable UI components.
package widgets

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Node represents a tree node.
type Node struct {
	Label    string
	Kind     string
	ID       string
	Children []*Node
	Expanded bool
	parent   *Node
}

// TreeModel is a tree widget with keyboard navigation.
type TreeModel struct {
	Root    *Node
	Focused *Node
	flat    []*Node
}

// NewTree creates a new tree model.
func NewTree(root *Node) TreeModel {
	t := TreeModel{Root: root}
	t.rebuild()
	if len(t.flat) > 0 {
		t.Focused = t.flat[0]
	}
	return t
}

// Init implements tea.Model.
func (t TreeModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (t TreeModel) Update(msg tea.Msg) (TreeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			t.MoveUp()
		case "down", "j":
			t.MoveDown()
		case "e":
			t.Expand()
		case "c":
			t.Collapse()
		case "E":
			t.ExpandAll()
		case "C":
			t.CollapseAll()
		}
	}
	return t, nil
}

// View implements tea.Model.
func (t TreeModel) View() string {
	var b strings.Builder
	for _, n := range t.flat {
		depth := t.depth(n)
		prefix := strings.Repeat("  ", depth)

		icon := " "
		if len(n.Children) > 0 {
			if n.Expanded {
				icon = "▼"
			} else {
				icon = "▶"
			}
		}

		marker := " "
		if n == t.Focused {
			marker = ">"
		}

		b.WriteString(marker)
		b.WriteString(prefix)
		b.WriteString(icon)
		b.WriteString(" ")
		b.WriteString(n.Label)
		b.WriteString("\n")
	}
	return b.String()
}

// MoveUp moves focus to the previous visible node.
func (t *TreeModel) MoveUp() {
	if t.Focused == nil || len(t.flat) == 0 {
		return
	}
	for i, n := range t.flat {
		if n == t.Focused && i > 0 {
			t.Focused = t.flat[i-1]
			break
		}
	}
}

// MoveDown moves focus to the next visible node.
func (t *TreeModel) MoveDown() {
	if t.Focused == nil || len(t.flat) == 0 {
		return
	}
	for i, n := range t.flat {
		if n == t.Focused && i < len(t.flat)-1 {
			t.Focused = t.flat[i+1]
			break
		}
	}
}

// Expand expands the focused node.
func (t *TreeModel) Expand() {
	if t.Focused != nil && len(t.Focused.Children) > 0 {
		t.Focused.Expanded = true
		t.rebuild()
	}
}

// Collapse collapses the focused node.
func (t *TreeModel) Collapse() {
	if t.Focused != nil {
		t.Focused.Expanded = false
		t.rebuild()
	}
}

// ExpandAll expands all nodes.
func (t *TreeModel) ExpandAll() {
	t.expandRecursive(t.Root)
	t.rebuild()
}

func (t *TreeModel) expandRecursive(n *Node) {
	if n == nil {
		return
	}
	n.Expanded = true
	for _, child := range n.Children {
		t.expandRecursive(child)
	}
}

// CollapseAll collapses all nodes.
func (t *TreeModel) CollapseAll() {
	t.collapseRecursive(t.Root)
	t.rebuild()
}

func (t *TreeModel) collapseRecursive(n *Node) {
	if n == nil {
		return
	}
	n.Expanded = false
	for _, child := range n.Children {
		t.collapseRecursive(child)
	}
}

// rebuild flattens the tree into a visible node list.
func (t *TreeModel) rebuild() {
	t.flat = nil
	t.flattenRecursive(t.Root)
}

func (t *TreeModel) flattenRecursive(n *Node) {
	if n == nil {
		return
	}
	t.flat = append(t.flat, n)
	if n.Expanded {
		for _, child := range n.Children {
			t.flattenRecursive(child)
		}
	}
}

func (t *TreeModel) depth(n *Node) int {
	d := 0
	for p := n.parent; p != nil; p = p.parent {
		d++
	}
	return d
}

// AddChild adds a child to a node and sets parent link.
func (n *Node) AddChild(child *Node) {
	child.parent = n
	n.Children = append(n.Children, child)
}
