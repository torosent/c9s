package containers

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/state"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/skinx"
	"github.com/torosent/c9s/internal/ui/theme"
)

// Model represents the containers screen.
type Model struct {
	client      cli.Client
	clk         clock.Clock
	palette     theme.Palette
	caps        cli.Capabilities
	tbl         table.Model
	containers  []cli.Container
	marks       map[string]bool
	filter      string
	filterMode  bool
	keymap      *keymap.Map
	width       int
	height      int
	sortKey     string
	sortReverse bool
}

// New creates a new containers screen.
func New(client cli.Client, clk clock.Clock, p theme.Palette) *Model {
	// Build keymap with screen-specific bindings overlaid on defaults
	km := keymap.Default()

	// Add container-specific bindings
	km.Add("shell", keymap.Binding{
		Keys:        []string{"s"},
		Help:        "Shell",
		Description: "Open shell in container",
	})
	km.Add("logs", keymap.Binding{
		Keys:        []string{"l", "enter"},
		Help:        "Logs",
		Description: "Tail container logs",
	})
	km.Add("inspect", keymap.Binding{
		Keys:        []string{"d"},
		Help:        "Details",
		Description: "Inspect container JSON",
	})
	km.Add("stop", keymap.Binding{
		Keys:        []string{"x"},
		Help:        "Stop",
		Description: "Stop container",
	})
	km.Add("kill", keymap.Binding{
		Keys:        []string{"shift+k", "K"},
		Help:        "Kill",
		Description: "Kill container",
	})
	km.Add("restart", keymap.Binding{
		Keys:        []string{"shift+r", "R"},
		Help:        "Restart",
		Description: "Restart container",
	})
	km.Add("delete", keymap.Binding{
		Keys:        []string{"shift+d", "D"},
		Help:        "Delete",
		Description: "Delete container",
	})
	km.Add("pause", keymap.Binding{
		Keys:        []string{"p"},
		Help:        "Pause",
		Description: "Pause container",
	})
	km.Add("pin", keymap.Binding{
		Keys:        []string{"b"},
		Help:        "Bookmark",
		Description: "Pin container",
	})

	// Create table; column widths are recomputed in reflowColumns when window
	// size is known.
	columns := []table.Column{
		{Title: " ", Width: 2},
		{Title: "SHORT-ID", Width: 12},
		{Title: "IMAGE", Width: 30},
		{Title: "STATE", Width: 10},
		{Title: "UPTIME", Width: 10},
		{Title: "CPU", Width: 5},
		{Title: "MEM", Width: 8},
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	tbl.Focus()
	tbl.SetStyles(skinx.TableStyles(p))

	return &Model{
		client:  client,
		clk:     clk,
		palette: p,
		tbl:     tbl,
		marks:   make(map[string]bool),
		keymap:  km,
	}
}

// Init implements screens.Screen.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		state.MakeRefreshedCmd[cli.Container](
			context.Background(),
			func(ctx context.Context) ([]cli.Container, error) {
				return m.client.ListContainers(ctx, true)
			},
			cli.ResourceContainers,
		),
		state.TickCmd(2*time.Second, m.clk, cli.ResourceContainers),
		func() tea.Msg {
			caps, _ := m.client.Capabilities(context.Background())
			return capabilitiesMsg(caps)
		},
	)
}

// Update implements screens.Screen.
func (m *Model) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tbl.SetHeight(msg.Height - 5)
		m.reflowColumns()
		m.rebuildTable()

	case capabilitiesMsg:
		m.caps = cli.Capabilities(msg)

	case modals.ConfirmResultMsg:
		if !msg.Result.Confirmed {
			return m, nil
		}
		switch msg.Result.Tag {
		case "delete":
			cmds = append(cmds, m.performDelete())
		}

	case state.RefreshedMsg[cli.Container]:
		if msg.Resource != cli.ResourceContainers {
			break
		}
		m.containers = msg.Snapshot.Items
		m.rebuildTable()

	case state.TickMsg:
		if msg.Resource != cli.ResourceContainers {
			break
		}
		// Trigger another refresh and re-arm the tick
		cmds = append(cmds,
			state.MakeRefreshedCmd[cli.Container](
				context.Background(),
				func(ctx context.Context) ([]cli.Container, error) {
					return m.client.ListContainers(ctx, true)
				},
				cli.ResourceContainers,
			),
			state.TickCmd(2*time.Second, m.clk, cli.ResourceContainers),
		)

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			// Compute row index (assuming table starts at Y=3 after title+header)
			if msg.Y >= 3 {
				row := msg.Y - 3
				if row >= 0 && row < len(m.containers) {
					m.tbl.SetCursor(row)
				}
			}
		case tea.MouseButtonWheelUp:
			m.tbl.MoveUp(1)
		case tea.MouseButtonWheelDown:
			m.tbl.MoveDown(1)
		}

	case tea.KeyMsg:
		if m.filterMode {
			return m.handleFilterKey(msg)
		}

		// Handle global keys
		if m.keymap.Matches("escape", msg) {
			m.marks = make(map[string]bool)
			m.rebuildTable()
			return m, nil
		}

		if m.keymap.Matches("mark", msg) {
			m.toggleMark()
			m.rebuildTable()
			return m, nil
		}

		if m.keymap.Matches("mark_all", msg) {
			m.selectAll()
			m.rebuildTable()
			return m, nil
		}

		if m.keymap.Matches("refresh", msg) {
			cmd := state.MakeRefreshedCmd[cli.Container](
				context.Background(),
				func(ctx context.Context) ([]cli.Container, error) {
					return m.client.ListContainers(ctx, true)
				},
				cli.ResourceContainers,
			)
			return m, cmd
		}

		if m.keymap.Matches("filter", msg) {
			m.filterMode = true
			m.filter = ""
			return m, nil
		}

		// Per-row actions
		if m.keymap.Matches("inspect", msg) {
			return m, m.inspectFocused()
		}

		if m.keymap.Matches("stop", msg) {
			return m, m.stopSelected()
		}

		if m.keymap.Matches("kill", msg) {
			return m, m.killSelected()
		}

		if m.keymap.Matches("restart", msg) {
			return m, m.restartSelected()
		}

		if m.keymap.Matches("delete", msg) {
			return m, m.deleteSelected()
		}

		if m.keymap.Matches("pause", msg) {
			if !m.caps.Pause {
				cmd := func() tea.Msg {
					return screens.StatusMsg{Toast: "pause not supported by this CLI version"}
				}
				return m, cmd
			}
			return m, m.pauseSelected()
		}

		if m.keymap.Matches("shell", msg) {
			return m, m.openShell()
		}

		if m.keymap.Matches("logs", msg) {
			return m, m.openLogs()
		}

		// Enter on the focused row opens logs (k9s convention).
		if msg.Type == tea.KeyEnter {
			return m, m.openLogs()
		}

		if m.keymap.Matches("pin", msg) {
			return m, m.pinFocused()
		}

		// Let table handle navigation
		m.tbl, _ = m.tbl.Update(msg)
	}

	return m, tea.Batch(cmds...)
}

// View implements screens.Screen.
func (m *Model) View(width, height int) string {
	body := m.tbl.View()
	if m.filterMode {
		body = m.tbl.View() + "\n" + fmt.Sprintf("Filter: %s_", m.filter)
	}
	filter := "all"
	if m.filter != "" {
		filter = m.filter
	}
	return skinx.BorderedBox(m.palette, "Containers", filter, len(m.containers), width, height, body)
}

// Title implements screens.Screen.
func (m *Model) Title() string {
	return "Containers"
}

// Hotkeys implements screens.Screen.
func (m *Model) Hotkeys() *keymap.Map {
	// If pause is unsupported, annotate the binding
	if !m.caps.Pause {
		// Create a copy of the keymap to avoid mutating the shared one
		km := &keymap.Map{}
		for _, name := range m.keymap.Names() {
			if b, ok := m.keymap.Get(name); ok {
				if name == "pause" {
					// Annotate pause binding
					b.Description = b.Description + " (unsupported by this container version)"
				}
				km.Add(name, b)
			}
		}
		return km
	}
	return m.keymap
}

// Summary implements screens.Screen.
func (m *Model) Summary() string {
	total := len(m.containers)
	running := 0
	exited := 0
	for _, c := range m.containers {
		switch c.Status {
		case "running":
			running++
		case "exited":
			exited++
		}
	}

	markedCount := len(m.marks)
	if markedCount > 0 {
		return fmt.Sprintf("%d items · %d running, %d exited · %d selected", total, running, exited, markedCount)
	}
	return fmt.Sprintf("%d items · %d running, %d exited", total, running, exited)
}

// rebuildTable rebuilds the table rows from the current container list.
func (m *Model) rebuildTable() {
	m.sortContainers()

	statusGlyph := func(s string) string {
		switch strings.ToLower(s) {
		case "running":
			return "● " + s
		case "stopped", "exited":
			return "○ " + s
		case "paused":
			return "⏸ " + s
		case "stopping":
			return "× " + s
		case "created":
			return "◌ " + s
		}
		return "  " + s
	}

	var rows []table.Row
	for _, c := range m.containers {
		if m.filter != "" {
			searchText := strings.ToLower(c.Image + " " + c.ShortID)
			if !strings.Contains(searchText, strings.ToLower(m.filter)) {
				continue
			}
		}

		mark := " "
		if m.marks[c.ID] {
			mark = "✓"
		}

		// Leading space on every column except the mark gives visual gap
		// between cells (bubbles/table has no built-in inter-column pad).
		row := table.Row{
			mark,
			" " + formatShortID(c.ID),
			" " + c.Image,
			" " + statusGlyph(c.Status),
			" " + formatUptime(c.Uptime),
			" " + fmt.Sprintf("%d", c.CPU),
			" " + formatBytes(c.MemBytes),
		}
		rows = append(rows, row)
	}
	m.tbl.SetRows(rows)
}

// reflowColumns recomputes column widths so the table fills the screen.
func (m *Model) reflowColumns() {
	if m.width <= 0 {
		return
	}
	markW := 2
	// Wider column allotments + bigger inter-column gap so cells don't crash
	// into each other (e.g. IMAGE column was running into STATE because the
	// glyph "○ stopped" plus whitespace was wider than the STATE allotment).
	stateW, uptimeW, cpuW, memW := 14, 12, 6, 10
	const padding = 8
	// Account for the bordered box (4 cols of border+pad) so we don't overflow.
	avail := m.width - 4
	remaining := avail - markW - stateW - uptimeW - cpuW - memW - padding
	if remaining < 30 {
		remaining = 30
	}
	shortIDW := 14
	imageW := remaining - shortIDW
	if imageW < 20 {
		imageW = 20
	}

	cols := []table.Column{
		{Title: " ", Width: markW},
		{Title: " SHORT-ID", Width: shortIDW},
		{Title: " IMAGE", Width: imageW},
		{Title: " STATE", Width: stateW},
		{Title: " UPTIME", Width: uptimeW},
		{Title: " CPU", Width: cpuW},
		{Title: " MEM", Width: memW},
	}
	m.tbl.SetColumns(cols)
	m.tbl.SetWidth(avail)
}

// toggleMark toggles the mark on the focused row.
func (m *Model) toggleMark() {
	c := m.focusedContainer()
	if c == nil {
		return
	}
	if m.marks[c.ID] {
		delete(m.marks, c.ID)
	} else {
		m.marks[c.ID] = true
	}
}

// selectAll marks all visible containers.
func (m *Model) selectAll() {
	for _, c := range m.containers {
		if m.filter != "" {
			searchText := strings.ToLower(c.Image + " " + c.ShortID)
			if !strings.Contains(searchText, strings.ToLower(m.filter)) {
				continue
			}
		}
		m.marks[c.ID] = true
	}
}

// focusedContainer returns the currently focused container.
func (m *Model) focusedContainer() *cli.Container {
	idx := m.tbl.Cursor()
	if idx < 0 || idx >= len(m.containers) {
		return nil
	}

	// Find the container by matching the focused row
	visibleContainers := m.visibleContainers()
	if idx >= len(visibleContainers) {
		return nil
	}
	return &visibleContainers[idx]
}

// visibleContainers returns the list of containers after filtering.
func (m *Model) visibleContainers() []cli.Container {
	if m.filter == "" {
		return m.containers
	}

	var visible []cli.Container
	for _, c := range m.containers {
		searchText := strings.ToLower(c.Image + " " + c.ShortID)
		if strings.Contains(searchText, strings.ToLower(m.filter)) {
			visible = append(visible, c)
		}
	}
	return visible
}

// targetIDs returns the list of IDs to operate on (focused or marked).
func (m *Model) targetIDs() []string {
	if len(m.marks) > 0 {
		ids := make([]string, 0, len(m.marks))
		for id := range m.marks {
			ids = append(ids, id)
		}
		return ids
	}

	c := m.focusedContainer()
	if c == nil {
		return nil
	}
	return []string{c.ID}
}

// inspectFocused opens the inspect modal for the focused container.
func (m *Model) inspectFocused() tea.Cmd {
	c := m.focusedContainer()
	if c == nil {
		return nil
	}

	return func() tea.Msg {
		ctx := context.Background()
		jsonBytes, err := m.client.InspectContainer(ctx, c.ID)
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("inspect failed: %v", err)}
		}
		return screens.OpenModalMsg{
			Modal: modals.NewInspect(fmt.Sprintf("Container %s", c.ShortID), jsonBytes, m.palette),
		}
	}
}

// stopSelected stops the targeted containers.
func (m *Model) stopSelected() tea.Cmd {
	ids := m.targetIDs()
	if len(ids) == 0 {
		return nil
	}

	var cmds []tea.Cmd
	for _, id := range ids {
		id := id // capture
		cmds = append(cmds, func() tea.Msg {
			ctx := context.Background()
			err := m.client.StopContainer(ctx, id)
			if err != nil {
				return screens.StatusMsg{Toast: fmt.Sprintf("stop %s failed: %v", id, err)}
			}
			return nil
		})
	}
	return tea.Batch(cmds...)
}

// killSelected kills the targeted containers.
func (m *Model) killSelected() tea.Cmd {
	ids := m.targetIDs()
	if len(ids) == 0 {
		return nil
	}

	var cmds []tea.Cmd
	for _, id := range ids {
		id := id
		cmds = append(cmds, func() tea.Msg {
			ctx := context.Background()
			err := m.client.KillContainer(ctx, id)
			if err != nil {
				return screens.StatusMsg{Toast: fmt.Sprintf("kill %s failed: %v", id, err)}
			}
			return nil
		})
	}
	return tea.Batch(cmds...)
}

// restartSelected restarts the targeted containers.
func (m *Model) restartSelected() tea.Cmd {
	ids := m.targetIDs()
	if len(ids) == 0 {
		return nil
	}

	var cmds []tea.Cmd
	for _, id := range ids {
		id := id
		cmds = append(cmds, func() tea.Msg {
			ctx := context.Background()
			err := m.client.RestartContainer(ctx, id)
			if err != nil {
				return screens.StatusMsg{Toast: fmt.Sprintf("restart %s failed: %v", id, err)}
			}
			return nil
		})
	}
	return tea.Batch(cmds...)
}

// deleteSelected opens a confirm modal and deletes on confirmation.
func (m *Model) deleteSelected() tea.Cmd {
	ids := m.targetIDs()
	if len(ids) == 0 {
		return nil
	}

	lines := make([]string, len(ids))
	for i, id := range ids {
		lines[i] = formatShortID(id)
	}

	return func() tea.Msg {
		return screens.OpenModalMsg{
			Modal: modals.NewConfirm(
				"Delete containers",
				"This will permanently remove:",
				lines,
				"delete",
				m.palette,
			),
		}
	}
}

// pauseSelected pauses the targeted containers.
func (m *Model) pauseSelected() tea.Cmd {
	ids := m.targetIDs()
	if len(ids) == 0 {
		return nil
	}

	var cmds []tea.Cmd
	for _, id := range ids {
		id := id
		cmds = append(cmds, func() tea.Msg {
			ctx := context.Background()
			err := m.client.PauseContainer(ctx, id)
			if err != nil {
				return screens.StatusMsg{Toast: fmt.Sprintf("pause %s failed: %v", id, err)}
			}
			return nil
		})
	}
	return tea.Batch(cmds...)
}

// openShell emits a SuspendShellMsg for the focused container.
func (m *Model) openShell() tea.Cmd {
	c := m.focusedContainer()
	if c == nil {
		return nil
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	return func() tea.Msg {
		return screens.SuspendShellMsg{
			ID:    c.ID,
			Shell: shell,
		}
	}
}

// openLogs opens the log viewer modal for the focused container (or all
// marked containers if any are marked).
func (m *Model) openLogs() tea.Cmd {
	ids := m.targetIDs()
	if len(ids) == 0 {
		return nil
	}
	// Build a {ID -> ShortID} lookup once
	display := make(map[string]string, len(m.containers))
	for _, c := range m.containers {
		display[c.ID] = formatShortID(c.ID)
		if display[c.ID] == "" {
			display[c.ID] = c.ID
		}
	}

	client := m.client
	return func() tea.Msg {
		ctx := context.Background()
		var sources []modals.LogSource
		for i, id := range ids {
			stream, err := client.StreamLogs(ctx, id, true)
			if err != nil {
				return screens.StatusMsg{Toast: fmt.Sprintf("logs %s failed: %v", id, err)}
			}
			name := display[id]
			if name == "" {
				name = id
			}
			sources = append(sources, modals.LogSource{
				Name:       name,
				ColorIndex: i,
				Stream:     stream,
			})
		}
		return screens.OpenModalMsg{Modal: modals.NewLogViewerWithPalette(sources, m.palette)}
	}
}

// pinFocused bookmarks the focused container into the :pinned screen.
func (m *Model) pinFocused() tea.Cmd {
	c := m.focusedContainer()
	if c == nil {
		return nil
	}
	id, name := c.ID, formatShortID(c.ID)
	if name == "" {
		name = c.ID
	}
	image := c.Image
	return func() tea.Msg {
		return screens.PinMsg{
			Resource: "containers",
			ID:       id,
			Display:  fmt.Sprintf("%s (%s)", name, image),
		}
	}
}

// performDelete actually deletes the targeted containers after confirmation.
func (m *Model) performDelete() tea.Cmd {
	ids := m.targetIDs()
	if len(ids) == 0 {
		return nil
	}
	client := m.client
	marks := m.marks
	var cmds []tea.Cmd
	for _, id := range ids {
		id := id
		cmds = append(cmds, func() tea.Msg {
			ctx := context.Background()
			if err := client.DeleteContainer(ctx, id); err != nil {
				return screens.StatusMsg{Toast: fmt.Sprintf("delete %s failed: %v", id, err)}
			}
			return screens.StatusMsg{Toast: fmt.Sprintf("deleted %s", formatShortID(id))}
		})
	}
	// Refresh after delete
	cmds = append(cmds, state.MakeRefreshedCmd[cli.Container](
		context.Background(),
		func(ctx context.Context) ([]cli.Container, error) {
			return client.ListContainers(ctx, true)
		},
		cli.ResourceContainers,
	))
	// Clear marks
	for k := range marks {
		delete(marks, k)
	}
	return tea.Batch(cmds...)
}

// handleFilterKey handles key input in filter mode.
func (m *Model) handleFilterKey(msg tea.KeyMsg) (screens.Screen, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.filterMode = false
		m.rebuildTable()
		return m, nil
	case tea.KeyEsc:
		m.filterMode = false
		m.filter = ""
		m.rebuildTable()
		return m, nil
	case tea.KeyBackspace:
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
		}
		return m, nil
	case tea.KeyRunes:
		m.filter += string(msg.Runes)
		return m, nil
	}
	return m, nil
}

// capabilitiesMsg is a message carrying CLI capabilities.
type capabilitiesMsg cli.Capabilities

// SortableColumns implements screens.Sortable.
func (m *Model) SortableColumns() []modals.SortColumn {
	return []modals.SortColumn{
		{Key: "id", Label: "ID"},
		{Key: "image", Label: "Image"},
		{Key: "status", Label: "Status"},
		{Key: "uptime", Label: "Uptime"},
		{Key: "cpu", Label: "CPU"},
		{Key: "mem", Label: "Memory"},
	}
}

// ApplySort implements screens.Sortable.
func (m *Model) ApplySort(key string, reverse bool) {
	m.sortKey = key
	m.sortReverse = reverse
	m.sortContainers()
	m.rebuildTable()
}

// sortContainers sorts the containers slice in place.
func (m *Model) sortContainers() {
	if m.sortKey == "" {
		return
	}

	sort.Slice(m.containers, func(i, j int) bool {
		less := false
		switch m.sortKey {
		case "id":
			less = m.containers[i].ID < m.containers[j].ID
		case "image":
			less = m.containers[i].Image < m.containers[j].Image
		case "status":
			less = m.containers[i].Status < m.containers[j].Status
		case "uptime":
			less = m.containers[i].Uptime < m.containers[j].Uptime
		case "cpu":
			less = m.containers[i].CPU < m.containers[j].CPU
		case "mem":
			less = m.containers[i].MemBytes < m.containers[j].MemBytes
		default:
			return false
		}
		if m.sortReverse {
			return !less
		}
		return less
	})
}
