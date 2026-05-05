// Package xray provides the :xray screen showing resource relationships.
package xray

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/theme"
	"github.com/torosent/c9s/internal/ui/widgets"
)

// Model represents the xray screen.
type Model struct {
	client  cli.Client
	palette theme.Palette
	tree    widgets.TreeModel
	keymap  *keymap.Map
	width   int
	height  int
}

// New creates a new xray screen.
func New(client cli.Client, p theme.Palette) *Model {
	km := keymap.Default()
	km.Add("jump", keymap.Binding{
		Keys:        []string{"enter"},
		Help:        "Jump",
		Description: "Jump to resource",
	})
	km.Add("expand", keymap.Binding{
		Keys:        []string{"e"},
		Help:        "Expand",
		Description: "Expand node",
	})
	km.Add("collapse", keymap.Binding{
		Keys:        []string{"c"},
		Help:        "Collapse",
		Description: "Collapse node",
	})

	root := &widgets.Node{Label: "Loading...", Kind: "root"}
	tree := widgets.NewTree(root)

	return &Model{
		client:  client,
		palette: p,
		tree:    tree,
		keymap:  km,
	}
}

// Init implements screens.Screen.
func (m *Model) Init() tea.Cmd {
	return m.buildTree()
}

func (m *Model) buildTree() tea.Cmd {
	return func() tea.Msg {
		root := &widgets.Node{Label: "Resources", Kind: "root", Expanded: true}

		// List containers
		containers, err := m.client.ListContainers(cli.DefaultCtx(), false)
		if err == nil {
			for _, c := range containers {
				cNode := &widgets.Node{
					Label:    fmt.Sprintf("container: %s (%s)", c.ShortID, c.Status),
					Kind:     "container",
					ID:       c.ID,
					Expanded: false,
				}
				root.AddChild(cNode)

				// Add image as child
				if c.Image != "" {
					imgNode := &widgets.Node{
						Label: fmt.Sprintf("image: %s", c.Image),
						Kind:  "image",
						ID:    c.Image,
					}
					cNode.AddChild(imgNode)
				}
			}
		}

		return TreeBuiltMsg{Root: root}
	}
}

// TreeBuiltMsg is sent when tree construction completes.
type TreeBuiltMsg struct {
	Root *widgets.Node
}

// Update implements screens.Screen.
func (m *Model) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case TreeBuiltMsg:
		m.tree = widgets.NewTree(msg.Root)

	case tea.KeyMsg:
		switch {
		case m.keymap.Matches("jump", msg):
			if m.tree.Focused != nil {
				node := m.tree.Focused
				// Emit jump message
				return m, func() tea.Msg {
					return JumpToResourceMsg{Kind: node.Kind, ID: node.ID}
				}
			}
		}

	case screens.PaletteChangedMsg:
		m.palette = msg.P

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	var cmd tea.Cmd
	m.tree, cmd = m.tree.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// JumpToResourceMsg requests switching to a resource's screen.
type JumpToResourceMsg struct {
	Kind string
	ID   string
}

// View implements screens.Screen.
func (m *Model) View(width, height int) string {
	if width != m.width || height != m.height {
		m.width = width
		m.height = height
	}
	return m.tree.View()
}

// Title implements screens.Screen.
func (m *Model) Title() string {
	return "xray"
}

// Hotkeys implements screens.Screen.
func (m *Model) Hotkeys() *keymap.Map {
	return m.keymap
}

// Summary implements screens.Screen.
func (m *Model) Summary() string {
	return "resource relationships"
}
