package images

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/state"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/skinx"
	"github.com/torosent/c9s/internal/ui/theme"
)

// Model represents the :images screen.
type Model struct {
	client      cli.Client
	clk         clock.Clock
	palette     theme.Palette
	tbl         table.Model
	images      []cli.Image
	marks       map[string]bool
	filter      string
	filterMode  bool
	keymap      *keymap.Map
	width       int
	height      int
	sortKey     string
	sortReverse bool
}

// RunRequestMsg is emitted when the user presses R on an image; the app
// is expected to open a Run form modal seeded with the given reference.
type RunRequestMsg struct {
	ImageRef string
}

// PushRequestMsg is emitted when the user presses P on an image; the app
// is expected to open a Push progress modal for the given reference.
type PushRequestMsg struct {
	ImageRef string
}

// TagRequestMsg is emitted when the user presses t on an image; the app
// is expected to prompt for the new tag.
type TagRequestMsg struct {
	Source string
}

// New creates a new Images screen model.
func New(client cli.Client, clk clock.Clock, p theme.Palette) *Model {
	km := keymap.Default()
	km.Add("inspect", keymap.Binding{
		Keys: []string{"d"}, Help: "Inspect", Description: "Inspect image JSON",
	})
	km.Add("tag", keymap.Binding{
		Keys: []string{"t"}, Help: "Tag", Description: "Tag image",
	})
	km.Add("push", keymap.Binding{
		Keys: []string{"P", "shift+p"}, Help: "Push", Description: "Push image to registry",
	})
	km.Add("delete", keymap.Binding{
		Keys: []string{"D", "shift+d"}, Help: "Delete", Description: "Delete image",
	})
	km.Add("run", keymap.Binding{
		Keys: []string{"enter", "R", "shift+r"}, Help: "Run", Description: "Run container from image",
	})
	km.Add("pin", keymap.Binding{
		Keys: []string{"b"}, Help: "Bookmark", Description: "Pin image",
	})

	columns := []table.Column{
		{Title: "REPOSITORY", Width: 32},
		{Title: "TAG", Width: 14},
		{Title: "ID", Width: 14},
		{Title: "CREATED", Width: 10},
		{Title: "SIZE", Width: 10},
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

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
		state.MakeRefreshedCmd[cli.Image](
			cli.DefaultCtx(),
			func(ctx context.Context) ([]cli.Image, error) {
				return m.client.ListImages(ctx)
			},
			cli.ResourceImages,
		),
		state.TickCmd(2*time.Second, m.clk, cli.ResourceImages),
	)
}

// Update implements screens.Screen.
func (m *Model) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case screens.PaletteChangedMsg:
		m.palette = msg.P
		m.tbl.SetStyles(skinx.TableStyles(msg.P))

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.height > 5 {
			m.tbl.SetHeight(m.height - 5)
		}

	case state.RefreshedMsg[cli.Image]:
		if msg.Resource != cli.ResourceImages {
			break
		}
		m.images = msg.Snapshot.Items
		m.rebuildTable()

	case state.TickMsg:
		if msg.Resource != cli.ResourceImages {
			break
		}
		cmds = append(cmds,
			state.MakeRefreshedCmd[cli.Image](
				cli.DefaultCtx(),
				func(ctx context.Context) ([]cli.Image, error) {
					return m.client.ListImages(ctx)
				},
				cli.ResourceImages,
			),
			state.TickCmd(2*time.Second, m.clk, cli.ResourceImages),
		)

	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			if msg.Y >= 3 {
				row := msg.Y - 3
				visible := m.visibleImages()
				if row >= 0 && row < len(visible) {
					m.tbl.SetCursor(row)
				}
			}
		}
	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelUp:
			m.tbl.MoveUp(1)
		case tea.MouseWheelDown:
			m.tbl.MoveDown(1)
		}

	case modals.ConfirmResultMsg:
		if msg.Result.Tag == "delete-images" && msg.Result.Confirmed {
			cmds = append(cmds, m.performDelete())
		}

	case modals.TextInputResultMsg:
		if strings.HasPrefix(msg.Result.Label, "tag-image:") {
			src := strings.TrimPrefix(msg.Result.Label, "tag-image:")
			cmds = append(cmds, m.performTag(src, msg.Result.Value))
		}

	case tea.KeyPressMsg:
		if m.filterMode {
			return m.handleFilterKey(msg)
		}

		if m.keymap.Matches("escape", msg) {
			m.marks = make(map[string]bool)
			return m, nil
		}
		if m.keymap.Matches("mark", msg) {
			m.toggleMark()
			return m, nil
		}
		if m.keymap.Matches("mark_all", msg) {
			m.selectAll()
			return m, nil
		}
		if m.keymap.Matches("refresh", msg) {
			return m, state.MakeRefreshedCmd[cli.Image](
				cli.DefaultCtx(),
				func(ctx context.Context) ([]cli.Image, error) {
					return m.client.ListImages(ctx)
				},
				cli.ResourceImages,
			)
		}
		if m.keymap.Matches("filter", msg) {
			m.filterMode = true
			m.filter = ""
			return m, nil
		}

		if m.keymap.Matches("inspect", msg) {
			return m, m.inspectFocused()
		}
		if m.keymap.Matches("tag", msg) {
			return m, m.requestTag()
		}
		if m.keymap.Matches("push", msg) {
			return m, m.requestPush()
		}
		if m.keymap.Matches("delete", msg) {
			return m, m.deleteSelected()
		}
		if m.keymap.Matches("run", msg) {
			return m, m.requestRun()
		}

		var cmd tea.Cmd
		m.tbl, cmd = m.tbl.Update(msg)
		cmds = append(cmds, cmd)
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
	return skinx.BorderedBox(m.palette, "Images", filter, len(m.images), width, height, body)
}

// Title implements screens.Screen.
func (m *Model) Title() string { return "Images" }

// Hotkeys implements screens.Screen.
func (m *Model) Hotkeys() *keymap.Map { return m.keymap }

// Summary implements screens.Screen.
func (m *Model) Summary() string {
	total := len(m.images)
	marked := len(m.marks)
	if marked > 0 {
		return fmt.Sprintf("%d items · %d selected", total, marked)
	}
	return fmt.Sprintf("%d items", total)
}

func (m *Model) rebuildTable() {
	m.sortImages()
	now := m.clk.Now()
	visible := m.visibleImages()
	rows := make([]table.Row, 0, len(visible))
	for _, img := range visible {
		repo := img.Repository
		if repo == "" {
			repo = "<none>"
		}
		tag := img.Tag
		if tag == "" {
			tag = "<none>"
		}
		rows = append(rows, table.Row{
			repo,
			tag,
			formatShortID(img.ID),
			formatCreated(img.Created, now),
			formatBytes(img.SizeBytes),
		})
	}
	m.tbl.SetRows(rows)
}

func (m *Model) visibleImages() []cli.Image {
	if m.filter == "" {
		return m.images
	}
	needle := strings.ToLower(m.filter)
	visible := make([]cli.Image, 0, len(m.images))
	for _, img := range m.images {
		hay := strings.ToLower(img.Repository + " " + img.Tag + " " + img.Reference + " " + img.ShortID)
		if strings.Contains(hay, needle) {
			visible = append(visible, img)
		}
	}
	return visible
}

func (m *Model) toggleMark() {
	img := m.focusedImage()
	if img == nil {
		return
	}
	if m.marks[img.ID] {
		delete(m.marks, img.ID)
		return
	}
	m.marks[img.ID] = true
}

func (m *Model) selectAll() {
	for _, img := range m.visibleImages() {
		m.marks[img.ID] = true
	}
}

func (m *Model) focusedImage() *cli.Image {
	idx := m.tbl.Cursor()
	visible := m.visibleImages()
	if idx < 0 || idx >= len(visible) {
		return nil
	}
	return &visible[idx]
}

func (m *Model) targetImages() []cli.Image {
	if len(m.marks) == 0 {
		if img := m.focusedImage(); img != nil {
			return []cli.Image{*img}
		}
		return nil
	}
	visible := m.visibleImages()
	out := make([]cli.Image, 0, len(m.marks))
	for _, img := range visible {
		if m.marks[img.ID] {
			out = append(out, img)
		}
	}
	return out
}

func (m *Model) inspectFocused() tea.Cmd {
	img := m.focusedImage()
	if img == nil {
		return nil
	}
	id := img.ID
	short := img.ShortID
	return func() tea.Msg {
		raw, err := m.client.InspectImage(cli.DefaultCtx(), id)
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("inspect %s failed: %v", short, err)}
		}
		return screens.OpenModalMsg{Modal: modals.NewInspect(
			fmt.Sprintf("Image %s", short), raw, m.palette,
		)}
	}
}

func (m *Model) requestTag() tea.Cmd {
	img := m.focusedImage()
	if img == nil {
		return nil
	}
	src := img.Reference
	if src == "" {
		src = img.ID
	}
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewTextInput(
			"tag-image:"+src,
			fmt.Sprintf("New tag for %s:", src),
			"",
			m.palette,
		)}
	}
}

func (m *Model) requestPush() tea.Cmd {
	img := m.focusedImage()
	if img == nil {
		return nil
	}
	ref := img.Reference
	if ref == "" {
		ref = img.ID
	}
	return func() tea.Msg {
		return PushRequestMsg{ImageRef: ref}
	}
}

func (m *Model) requestRun() tea.Cmd {
	img := m.focusedImage()
	if img == nil {
		return nil
	}
	ref := img.Reference
	if ref == "" {
		ref = img.ID
	}
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewRunForm(ref, m.palette)}
	}
}

func (m *Model) deleteSelected() tea.Cmd {
	imgs := m.targetImages()
	if len(imgs) == 0 {
		return nil
	}
	lines := make([]string, len(imgs))
	for i, img := range imgs {
		ref := img.Reference
		if ref == "" {
			ref = img.ShortID
		}
		lines[i] = ref
	}
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewConfirm(
			"Delete images",
			"This will permanently remove:",
			lines,
			"delete-images",
			m.palette,
		)}
	}
}

func (m *Model) performTag(src, dst string) tea.Cmd {
	dst = strings.TrimSpace(dst)
	if dst == "" {
		return func() tea.Msg {
			return screens.StatusMsg{Toast: "tag: destination is empty"}
		}
	}
	return func() tea.Msg {
		err := m.client.TagImage(cli.DefaultCtx(), src, dst)
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("tag %s → %s failed: %v", src, dst, err)}
		}
		return screens.StatusMsg{Toast: fmt.Sprintf("tagged %s as %s", src, dst)}
	}
}

func (m *Model) performDelete() tea.Cmd {
	imgs := m.targetImages()
	if len(imgs) == 0 {
		return nil
	}
	cmds := make([]tea.Cmd, 0, len(imgs))
	for _, img := range imgs {
		id := img.ID
		short := img.ShortID
		cmds = append(cmds, func() tea.Msg {
			err := m.client.DeleteImage(cli.DefaultCtx(), id)
			if err != nil {
				return screens.StatusMsg{Toast: fmt.Sprintf("delete %s failed: %v", short, err)}
			}
			return screens.StatusMsg{Toast: fmt.Sprintf("deleted %s", short)}
		})
	}
	m.marks = make(map[string]bool)
	return tea.Batch(cmds...)
}

func (m *Model) handleFilterKey(msg tea.KeyPressMsg) (screens.Screen, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.filterMode = false
		m.rebuildTable()
		return m, nil
	case "esc":
		m.filterMode = false
		m.filter = ""
		m.rebuildTable()
		return m, nil
	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.rebuildTable()
		}
		return m, nil
	default:
		if msg.Text != "" {
			m.filter += msg.Text
			m.rebuildTable()
			return m, nil
		}
	}
	return m, nil
}

// SortableColumns implements screens.Sortable.
func (m *Model) SortableColumns() []modals.SortColumn {
	return []modals.SortColumn{
		{Key: "id", Label: "ID"},
		{Key: "repository", Label: "Repository"},
		{Key: "tag", Label: "Tag"},
		{Key: "created", Label: "Created"},
		{Key: "size", Label: "Size"},
	}
}

// ApplySort implements screens.Sortable.
func (m *Model) ApplySort(key string, reverse bool) {
	m.sortKey = key
	m.sortReverse = reverse
	m.sortImages()
	m.rebuildTable()
}

// sortImages sorts the images slice in place.
func (m *Model) sortImages() {
	if m.sortKey == "" {
		return
	}

	sort.Slice(m.images, func(i, j int) bool {
		less := false
		switch m.sortKey {
		case "id":
			less = m.images[i].ID < m.images[j].ID
		case "repository":
			less = m.images[i].Repository < m.images[j].Repository
		case "tag":
			less = m.images[i].Tag < m.images[j].Tag
		case "created":
			less = m.images[i].Created.Before(m.images[j].Created)
		case "size":
			less = m.images[i].SizeBytes < m.images[j].SizeBytes
		default:
			return false
		}
		if m.sortReverse {
			return !less
		}
		return less
	})
}
