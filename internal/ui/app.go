package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/cloud/acr"
	"github.com/torosent/c9s/internal/config"
	"github.com/torosent/c9s/internal/dockershim"
	"github.com/torosent/c9s/internal/jobs"
	"github.com/torosent/c9s/internal/log"
	"github.com/torosent/c9s/internal/pinned"
	"github.com/torosent/c9s/internal/state"
	"github.com/torosent/c9s/internal/ui/breadcrumbs"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/palette"
	"github.com/torosent/c9s/internal/ui/screens"
	builderscreen "github.com/torosent/c9s/internal/ui/screens/builder"
	containerscreen "github.com/torosent/c9s/internal/ui/screens/containers"
	errorsscreen "github.com/torosent/c9s/internal/ui/screens/errors"
	imagesscreen "github.com/torosent/c9s/internal/ui/screens/images"
	jobsscreen "github.com/torosent/c9s/internal/ui/screens/jobs"
	networksscreen "github.com/torosent/c9s/internal/ui/screens/networks"
	pinnedscreen "github.com/torosent/c9s/internal/ui/screens/pinned"
	pulsesscreen "github.com/torosent/c9s/internal/ui/screens/pulses"
	registryscreen "github.com/torosent/c9s/internal/ui/screens/registry"
	systemscreen "github.com/torosent/c9s/internal/ui/screens/system"
	volumesscreen "github.com/torosent/c9s/internal/ui/screens/volumes"
	xrayscreen "github.com/torosent/c9s/internal/ui/screens/xray"
	"github.com/torosent/c9s/internal/ui/theme"
	"github.com/torosent/c9s/internal/version"
)

// Model is the root tea.Model for c9s.
type Model struct {
	client   cli.Client
	clk      clock.Clock
	palette  theme.Palette
	errorLog *log.Logger
	pinStore *pinned.Store
	history  *palette.History
	crumbs   *breadcrumbs.Trail
	jobMgr   *jobs.Manager

	width, height int

	showSplash bool
	splash     SplashModel
	statusBar  StatusBar

	screens       map[string]screens.Screen
	active        string
	stack         modals.Stack
	headerVisible bool

	caps    cli.Capabilities
	capsErr error

	cmdActive bool
	cmdBuf    string
	aliases   map[string]string // command aliases from config
	readonly  bool              // read-only mode disables destructive operations
	skinName  string            // currently loaded skin name (for header display)

	toast string
}

type capabilitiesMsg struct {
	caps cli.Capabilities
	err  error
}

// acrLoginMsg is delivered after a `:acr-login` invocation completes,
// carrying either an error (FetchToken or RegistryLogin failed) or
// success (host is set, err is nil) so Update can render a toast.
type acrLoginMsg struct {
	host string
	err  error
}

// shellExecDoneMsg is emitted after tea.ExecProcess returns from a
// SuspendShellMsg. Carries an optional toast (set when the exec
// failed) and triggers a fresh WindowSizeMsg so altscreen is fully
// repainted — without that, the post-exec frame can render on top of
// stale cells and leave the screen looking glitched.
type shellExecDoneMsg struct {
	toast string
}

// NewApp constructs the root model.
func NewApp(client cli.Client, clk clock.Clock, p theme.Palette, cfg config.Config) Model {
	// Set up data directories
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".local", "share")
	}
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		home, _ := os.UserHomeDir()
		stateDir = filepath.Join(home, ".local", "state")
	}

	logDir := filepath.Join(dataDir, "c9s")
	pinnedPath := filepath.Join(stateDir, "c9s", "pinned.toml")
	historyPath := filepath.Join(stateDir, "c9s", "history")

	// Initialize error logger
	errorLog, err := log.New(logDir, clk)
	if err != nil {
		// Non-fatal - continue without error logging
		errorLog = nil
	}

	// Initialize pinned store
	pinStore, err := pinned.Load(pinnedPath)
	if err != nil {
		// Non-fatal - continue without pinned store
		pinStore = nil
	}

	// Initialize command history
	hist, err := palette.Load(historyPath)
	if err != nil {
		// Non-fatal - continue without history
		hist = nil
	}

	// Initialize jobs manager
	jobMgr := jobs.New(clk)

	// Register screens. Each screen is created once and re-initialized
	// when it becomes the active screen.
	screenMap := make(map[string]screens.Screen)
	screenMap["containers"] = containerscreen.New(client, clk, p)
	screenMap["images"] = imagesscreen.New(client, clk, p)
	screenMap["volumes"] = volumesscreen.New(client, clk, p)
	screenMap["networks"] = networksscreen.New(client, clk, p)
	screenMap["builder"] = builderscreen.New(client, clk, p)
	screenMap["registry"] = registryscreen.New(client, clk, p)
	screenMap["system"] = systemscreen.NewServices(client, clk, p)
	screenMap["df"] = systemscreen.NewDF(client, clk, p)
	screenMap["dns"] = systemscreen.NewDNS(client, clk, p)
	screenMap["property"] = systemscreen.NewProperty(client, clk, p)
	screenMap["kernel"] = systemscreen.NewKernel(client, clk, p)
	screenMap["syslogs"] = systemscreen.NewLogs(client, clk, p)
	screenMap["errors"] = errorsscreen.New(logDir, clk, p)
	screenMap["pinned"] = pinnedscreen.New(pinStore, p)
	screenMap["xray"] = xrayscreen.New(client, p)
	screenMap["pulses"] = pulsesscreen.New(client, clk, p)
	screenMap["jobs"] = jobsscreen.New(jobMgr, clk, p)

	return Model{
		client:        client,
		clk:           clk,
		palette:       p,
		errorLog:      errorLog,
		pinStore:      pinStore,
		history:       hist,
		crumbs:        breadcrumbs.New(),
		jobMgr:        jobMgr,
		showSplash:    true,
		splash:        NewSplash(p, "c9s — Apple Containers TUI"),
		statusBar:     NewStatusBar(p),
		screens:       screenMap,
		active:        "containers",
		headerVisible: true,
		aliases:       cfg.Aliases,
		readonly:      cfg.UI.ReadOnly,
		skinName:      "default",
	}
}

// Init kicks off the capabilities probe and initializes the active screen.
func (m Model) Init() tea.Cmd {
	// Initialize breadcrumbs with the initial screen
	if scr, ok := m.screens[m.active]; ok {
		m.crumbs.Push(breadcrumbs.Crumb{Label: scr.Title(), Screen: m.active})
	}
	return tea.Batch(
		m.splash.Init(),
		m.probeCmd(),
		m.screens["containers"].Init(),
	)
}

func (m Model) probeCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		caps, err := m.client.Capabilities(ctx)
		return capabilitiesMsg{caps: caps, err: err}
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.showSplash {
			var cmd tea.Cmd
			m.splash, cmd = m.splash.Update(msg)
			return m, cmd
		}
		var cmds []tea.Cmd
		// Always propagate to the active screen so it can reflow.
		if scr, ok := m.screens[m.active]; ok {
			newScr, cmd := scr.Update(msg)
			m.screens[m.active] = newScr
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		// Also propagate to the top modal if open, so its viewport resizes.
		if !m.stack.Empty() {
			modal := m.stack.Top()
			newModal, cmd := modal.Update(msg)
			m.stack.Pop()
			m.stack.Push(newModal)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)

	case capabilitiesMsg:
		m.caps = msg.caps
		m.capsErr = msg.err
		if msg.err != nil {
			m.toast = "could not probe `container`: " + msg.err.Error()
		}
		return m, nil

	case acrLoginMsg:
		if msg.err != nil {
			m.toast = fmt.Sprintf("acr-login %s failed", msg.host)
			m.logError("registry.acr-login", msg.host, fmt.Sprintf("acr-login failed: %v", msg.err), msg.err.Error())
			modal := modals.NewInfo(
				fmt.Sprintf("ACR login failed: %s", msg.host),
				[]string{
					msg.err.Error(),
					"",
					"Common fixes:",
					"  • Run `az login` first if Azure CLI isn't authenticated.",
					"  • Verify the registry name (or pass full myreg.azurecr.io).",
					"  • Check `az acr show --name <reg>` resolves the registry.",
				},
				modals.InfoError,
				m.palette,
			)
			m.stack.Push(modal)
			return m, modal.Init()
		}
		m.toast = fmt.Sprintf("logged in to %s", msg.host)
		modal := modals.NewInfo(
			"ACR login succeeded",
			[]string{
				fmt.Sprintf("Logged in to %s via Azure AD.", msg.host),
				"",
				"The token is stored by Apple's `container` for ~3 hours;",
				"re-run `:acr-login` when it expires. Existing pulls and",
				"pushes will use the credential automatically.",
			},
			modals.InfoOK,
			m.palette,
		)
		m.stack.Push(modal)
		return m, modal.Init()

	case SplashDoneMsg:
		m.showSplash = false
		return m, nil

	case screens.OpenModalMsg:
		m.stack.Push(msg.Modal)
		m.crumbs.Push(breadcrumbs.Crumb{Label: "modal:" + msg.Modal.Title()})
		return m, msg.Modal.Init()

	case modals.CloseModalMsg:
		m.stack.Pop()
		if m.crumbs.Len() > 1 {
			m.crumbs.Pop()
		}
		return m, nil

	case modals.SortPickedMsg:
		m.stack.Pop()
		// Apply sort to current screen if it implements Sortable
		if scr, ok := m.screens[m.active].(interface {
			ApplySort(key string, reverse bool)
		}); ok {
			scr.ApplySort(msg.Key, msg.Reverse)
		}
		return m, nil

	case modals.SkinPickedMsg:
		// Pop the picker modal + its breadcrumb, then apply via runCommand("skin <name>")
		m.stack.Pop()
		if m.crumbs.Len() > 1 {
			m.crumbs.Pop()
		}
		return m.runCommand("skin " + msg.Name)

	case modals.ConfirmResultMsg:
		// Forward the message to the active screen
		if scr, ok := m.screens[m.active]; ok {
			newScr, cmd := scr.Update(msg)
			m.screens[m.active] = newScr
			return m, cmd
		}
		return m, nil

	case modals.ShellPickedMsg:
		// The shell-picker modal batches ShellPickedMsg alongside
		// CloseModalMsg, but tea.Batch makes no ordering guarantees.
		// If we let this fall through to the catch-all routing
		// below, the message races CloseModalMsg: when ShellPickedMsg
		// arrives first the picker is still on the stack, the modal
		// receives the message, doesn't handle it, and the pick is
		// silently dropped — exactly the "I clicked bash and nothing
		// happened" symptom. Forward directly to the active screen,
		// matching the ConfirmResultMsg pattern above.
		if scr, ok := m.screens[m.active]; ok {
			newScr, cmd := scr.Update(msg)
			m.screens[m.active] = newScr
			return m, cmd
		}
		return m, nil

	case modals.LoginResultMsg, modals.LoginCancelledMsg,
		modals.TextInputResultMsg, modals.TextInputCancelledMsg:
		if scr, ok := m.screens[m.active]; ok {
			newScr, cmd := scr.Update(msg)
			m.screens[m.active] = newScr
			return m, cmd
		}
		return m, nil

	case modals.RunSubmittedMsg:
		// Open progress modal streaming the run.
		stream, err := m.client.RunContainer(context.Background(), msg.Opts)
		if err != nil {
			m.toast = fmt.Sprintf("run failed: %v", err)
			m.logError("container.run", msg.Opts.Image, fmt.Sprintf("run failed: %v", err), err.Error())
			return m, nil
		}
		modal := modals.NewProgressModel(jobs.KindBuild, msg.Opts.Image, stream, m.clk)
		// Wrap as a Modal — but ProgressModel is a tea.Model, not Modal.
		// We adapt via a small inline progress wrapper.
		m.stack.Push(progressModalWrap{p: modal})
		return m, modal.Init()

	case modals.BuildSubmittedMsg:
		stream, err := m.client.StreamBuild(context.Background(), msg.Opts)
		if err != nil {
			m.toast = fmt.Sprintf("build failed: %v", err)
			return m, nil
		}
		modal := modals.NewProgressModel(jobs.KindBuild, msg.Opts.Tag, stream, m.clk)
		m.stack.Push(progressModalWrap{p: modal})
		return m, modal.Init()

	case modals.RunCancelledMsg, modals.BuildCancelledMsg:
		m.toast = "cancelled"
		return m, nil

	case imagesscreen.PushRequestMsg:
		stream, err := m.client.StreamPush(context.Background(), msg.ImageRef)
		if err != nil {
			m.toast = fmt.Sprintf("push failed: %v", err)
			return m, nil
		}
		modal := modals.NewProgressModel(jobs.KindPush, msg.ImageRef, stream, m.clk)
		m.stack.Push(progressModalWrap{p: modal})
		return m, modal.Init()

	case screens.SuspendShellMsg:
		// Apple's `container exec` returns exit 0 EVEN WHEN THE SHELL
		// ISN'T INSTALLED — it writes the error to stderr (visible
		// for milliseconds before altscreen re-entry hides it) and
		// then exits cleanly. tea.ExecProcess sees a 0 exit code, so
		// we can't surface a useful error post-hoc. Probe first via
		// `container exec <id> test -x <shell>` (no -i/-t, so this
		// returns a real exit code) and toast immediately if the
		// shell isn't there. This is also why we ditched the host
		// $SHELL — /bin/zsh is rarely in a Linux container, and the
		// silent-failure mode left users staring at a "nothing
		// happened" screen.
		shortID := msg.ID
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}
		probeCtx, probeCancel := context.WithTimeout(context.Background(), 3*time.Second)
		// #nosec G204 -- ID/Shell originate from internal CLI snapshots and the modal's static option list (/bin/bash | /bin/sh), not user-supplied strings; binary path is the configured cli.Client.Bin().
		probe := exec.CommandContext(probeCtx, m.client.Bin(), "exec", msg.ID, "test", "-x", msg.Shell)
		probeErr := probe.Run()
		probeCancel()
		if probeErr != nil {
			m.toast = fmt.Sprintf("%s not available in %s — try the other shell", msg.Shell, shortID)
			return m, nil
		}

		// Shell exists. Run `container exec -it <id> <shell>` via
		// tea.ExecProcess (exits altscreen, runs the child, then
		// re-enters altscreen).
		// #nosec G204 -- ID/Shell originate from internal CLI snapshots and the modal's static option list (/bin/bash | /bin/sh), not user-supplied strings; binary path is the configured cli.Client.Bin().
		execCmd := exec.Command(m.client.Bin(), "exec", "-it", msg.ID, msg.Shell)
		execDone := tea.ExecProcess(execCmd, func(err error) tea.Msg {
			toast := ""
			if err != nil {
				toast = fmt.Sprintf("shell %s failed: %v", shortID, err)
			}
			return shellExecDoneMsg{toast: toast}
		})
		return m, execDone

	case shellExecDoneMsg:
		if msg.toast != "" {
			m.toast = msg.toast
		}
		// Force a full altscreen rebuild after exec returns. After
		// hours of trial — tea.ClearScreen alone, tea.WindowSize()
		// (async), and synthetic WindowSizeMsg with known dims all
		// failed to repaint cleanly in some terminals — the only
		// thing that consistently works is toggling altscreen off
		// then on. Bubbletea's RestoreTerminal calls
		// renderer.enterAltScreen() unconditionally if altscreen was
		// active, but that's idempotent — the altscreen is already
		// active so it's a no-op. Forcing ExitAltScreen first makes
		// the subsequent EnterAltScreen actually run the entry
		// sequence (\033[?1049h) and reset the buffer.
		//
		// Sequence (not Batch) so the toggle and reflow run in
		// strict order: exit → enter → reflow → re-init.
		width, height := m.width, m.height
		var initCmd tea.Cmd
		if scr, ok := m.screens[m.active]; ok {
			initCmd = scr.Init()
		}
		seq := []tea.Cmd{
			func() tea.Msg { return tea.ExitAltScreen() },
			func() tea.Msg { return tea.EnterAltScreen() },
			func() tea.Msg { return tea.ClearScreen() },
			func() tea.Msg { return tea.WindowSizeMsg{Width: width, Height: height} },
		}
		if initCmd != nil {
			seq = append(seq, initCmd)
		}
		return m, tea.Sequence(seq...)

	case screens.StatusMsg:
		m.toast = msg.Toast
		return m, nil

	case screens.PinMsg:
		if m.pinStore != nil {
			_ = m.pinStore.Pin(pinned.Pin{
				Resource: msg.Resource,
				ID:       msg.ID,
				Display:  msg.Display,
				Added:    m.clk.Now(),
			})
			m.toast = fmt.Sprintf("pinned %s", msg.Display)
		} else {
			m.toast = "pin store unavailable"
		}
		return m, nil

	case tea.KeyMsg:
		if m.showSplash {
			var cmd tea.Cmd
			m.splash, cmd = m.splash.Update(msg)
			return m, cmd
		}

		// If modal stack is non-empty, route to top modal
		if !m.stack.Empty() {
			modal := m.stack.Top()
			newModal, cmd := modal.Update(msg)
			m.stack.Pop()
			m.stack.Push(newModal)
			return m, cmd
		}

		// Handle colon palette
		if m.cmdActive {
			return m.handleCommandKey(msg)
		}

		// Check global keys
		globalMap := keymap.Default()
		if globalMap.Matches("palette", msg) {
			m.cmdActive = true
			return m, nil
		}
		if globalMap.Matches("interrupt", msg) {
			return m, tea.Quit
		}
		if globalMap.Matches("header_toggle", msg) {
			m.headerVisible = !m.headerVisible
			return m, nil
		}
		if globalMap.Matches("help", msg) {
			if scr, ok := m.screens[m.active]; ok {
				modal := modals.NewHelp(scr.Hotkeys(), scr.Title(), m.palette)
				m.stack.Push(modal)
				return m, modal.Init()
			}
		}

		// Handle Shift+S for sort picker
		if msg.String() == "S" {
			if sortable, ok := m.screens[m.active].(interface {
				SortableColumns() []modals.SortColumn
			}); ok {
				cols := sortable.SortableColumns()
				if len(cols) > 0 {
					modal := modals.NewSortPicker(cols, m.palette)
					m.stack.Push(modal)
					return m, modal.Init()
				}
			}
		}

		// Number shortcuts for direct screen access (k9s-style).
		if scr := numberShortcut(msg.String()); scr != "" {
			return m.runCommand(scr)
		}

		// Forward to active screen
		if scr, ok := m.screens[m.active]; ok {
			newScr, cmd := scr.Update(msg)
			m.screens[m.active] = newScr
			return m, cmd
		}
		return m, nil
	}

	// Forward other messages: prefer the top modal when one is open,
	// otherwise the active screen. This is essential for streaming modals
	// (log viewer, progress modal) that drive themselves via tea.Cmd
	// loops emitting their own private message types.
	//
	// Special-case state.TickMsg: ALWAYS forward to the active screen so
	// the screen's polling loop (which re-arms a new TickCmd in its
	// handler) keeps running while a modal is open. Otherwise the modal
	// swallows the tick, the underlying screen's data freezes, and after
	// the modal closes the user sees stale data.
	//
	// Forwarding also runs while the splash is showing: the active
	// screen's Init dispatches an immediate fetch (RefreshedMsg) and
	// arms its first TickCmd. If those messages were dropped during
	// splash, the screen would render empty until the user happened to
	// trigger another fetch (e.g., a screen switch), and—because
	// clock.Real().Tick() is one-shot via time.After—the polling loop
	// would never re-arm. Splash only intercepts keys/window-size in
	// the typed cases above.
	if _, isTick := msg.(state.TickMsg); isTick {
		if scr, ok := m.screens[m.active]; ok {
			newScr, cmd := scr.Update(msg)
			m.screens[m.active] = newScr
			return m, cmd
		}
	}
	if !m.stack.Empty() {
		modal := m.stack.Top()
		newModal, cmd := modal.Update(msg)
		m.stack.Pop()
		m.stack.Push(newModal)
		return m, cmd
	}
	if scr, ok := m.screens[m.active]; ok {
		newScr, cmd := scr.Update(msg)
		m.screens[m.active] = newScr
		return m, cmd
	}

	return m, nil
}

func (m Model) handleCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.cmdActive = false
		m.cmdBuf = ""
		if m.history != nil {
			m.history.Reset()
		}
		return m, nil
	case tea.KeyEnter:
		cmd := strings.TrimSpace(m.cmdBuf)
		m.cmdActive = false
		m.cmdBuf = ""
		if m.history != nil && cmd != "" {
			_ = m.history.Append(cmd)
			m.history.Reset()
		}
		return m.runCommand(cmd)
	case tea.KeyBackspace:
		if len(m.cmdBuf) > 0 {
			m.cmdBuf = m.cmdBuf[:len(m.cmdBuf)-1]
		}
		return m, nil
	case tea.KeyTab:
		// Autocomplete to the longest common prefix of matching commands.
		matches := palette.Match(m.cmdBuf, palette.Catalog())
		if len(matches) == 0 {
			return m, nil
		}
		if len(matches) == 1 {
			m.cmdBuf = matches[0].Name
			return m, nil
		}
		// Multiple matches → set buffer to longest common prefix
		lcp := matches[0].Name
		for _, c := range matches[1:] {
			lcp = commonPrefix(lcp, c.Name)
			if lcp == "" {
				break
			}
		}
		if len(lcp) > len(m.cmdBuf) {
			m.cmdBuf = lcp
		}
		return m, nil
	case tea.KeyUp:
		if m.history != nil {
			if prev := m.history.Up(); prev != "" {
				m.cmdBuf = prev
			}
		}
		return m, nil
	case tea.KeyDown:
		if m.history != nil {
			if next := m.history.Down(); next != "" {
				m.cmdBuf = next
			} else {
				m.cmdBuf = ""
			}
		}
		return m, nil
	case tea.KeySpace:
		m.cmdBuf += " "
		return m, nil
	case tea.KeyRunes:
		m.cmdBuf += string(msg.Runes)
		return m, nil
	}
	return m, nil
}

func commonPrefix(a, b string) string {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}
	return a[:n]
}

func (m Model) runCommand(cmd string) (tea.Model, tea.Cmd) {
	cmd = strings.TrimSpace(cmd)
	// Parse "<verb> <arg...>" so :tag/save/load/run/build/login/create can carry args
	verb, arg := cmd, ""
	if idx := strings.IndexAny(cmd, " \t"); idx >= 0 {
		verb = cmd[:idx]
		arg = strings.TrimSpace(cmd[idx+1:])
	}

	// Resolve alias first
	verb = config.ResolveAlias(verb, m.aliases)

	// Check readonly mode for destructive commands
	if m.readonly {
		destructive := map[string]bool{
			"delete": true, "prune": true, "stop": true, "kill": true,
			"pause": true, "unpause": true, "rm": true, "rmi": true,
			"start": true, "restart": true, "remove": true,
		}
		if destructive[verb] {
			m.toast = "Operation blocked: read-only mode"
			return m, nil
		}
	}

	// Screen-switching aliases. Top-level screens reset the trail rather than
	// accumulating crumbs; sub-screens (df/dns/property/kernel/syslogs, when
	// :system is the parent) and modals push and pop normally.
	subScreens := map[string]bool{
		"df": true, "dns": true, "property": true, "kernel": true, "syslogs": true,
	}
	switchTo := func(name string) (tea.Model, tea.Cmd) {
		m.active = name
		if scr, ok := m.screens[name]; ok {
			if subScreens[name] {
				m.crumbs.Push(breadcrumbs.Crumb{Label: scr.Title(), Screen: name})
			} else {
				m.crumbs.Replace(breadcrumbs.Crumb{Label: scr.Title(), Screen: name})
			}
			return m, scr.Init()
		}
		m.toast = fmt.Sprintf("screen %q not registered", name)
		return m, nil
	}

	switch verb {
	case "q", "quit", "exit":
		return m, tea.Quit
	case "containers", "c":
		return switchTo("containers")
	case "images", "i":
		return switchTo("images")
	case "volumes", "v":
		return switchTo("volumes")
	case "networks", "n":
		return switchTo("networks")
	case "builder", "b":
		return switchTo("builder")
	case "registry", "reg":
		return switchTo("registry")
	case "system", "sys":
		return switchTo("system")
	case "df":
		return switchTo("df")
	case "dns":
		return switchTo("dns")
	case "property":
		return switchTo("property")
	case "kernel":
		return switchTo("kernel")
	case "logs":
		return switchTo("syslogs")
	case "errors":
		return switchTo("errors")
	case "pinned":
		return switchTo("pinned")
	case "xray":
		return switchTo("xray")
	case "pulses":
		return switchTo("pulses")

	case "help", "?":
		if scr, ok := m.screens[m.active]; ok {
			modal := modals.NewHelp(scr.Hotkeys(), scr.Title(), m.palette)
			m.stack.Push(modal)
			return m, modal.Init()
		}
		return m, nil
	case "run":
		modal := modals.NewRunForm(arg, m.palette)
		m.stack.Push(modal)
		return m, modal.Init()
	case "build":
		modal := modals.NewBuildForm(arg, m.palette)
		m.stack.Push(modal)
		return m, modal.Init()
	case "login":
		modal := modals.NewLogin(arg, m.palette)
		m.stack.Push(modal)
		return m, modal.Init()

	case "acr-login":
		if arg == "" {
			m.toast = "usage: :acr-login <registry> · accepts 'myreg' or 'myreg.azurecr.io'"
			return m, nil
		}
		host := acr.Hostname(arg)
		m.toast = fmt.Sprintf("fetching ACR token for %s …", host)
		client := m.client
		registry := arg
		return m, func() tea.Msg {
			// 30 s covers slow `az` token-cache rehydration. If `az` would
			// need to prompt the user for an interactive `az login`, that
			// would exceed this and surface as a context-deadline error,
			// telling the user to run `az login` first.
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			token, err := acr.FetchToken(ctx, registry)
			if err != nil {
				return acrLoginMsg{host: host, err: err}
			}
			if err := client.RegistryLogin(ctx, host, acr.AnonymousUser, token); err != nil {
				return acrLoginMsg{host: host, err: fmt.Errorf("container registry login: %w", err)}
			}
			return acrLoginMsg{host: host}
		}

	case "install-docker-shim":
		path := arg
		if path == "" {
			def, err := dockershim.DefaultPath()
			if err != nil {
				m.toast = fmt.Sprintf("install-docker-shim: %v", err)
				return m, nil
			}
			path = def
		}
		// Resolve to absolute so the toast hint refers to a real
		// directory the user can put on PATH.
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		// Detect existing docker(s) so the modal can warn about PATH
		// precedence in the same breath as reporting the install.
		others, _ := dockershim.DetectExistingDocker(path)
		if err := dockershim.Install(path, false); err != nil {
			m.toast = "install-docker-shim failed"
			m.logError("dockershim.install", path, fmt.Sprintf("install failed: %v", err), err.Error())
			modal := modals.NewInfo(
				"Docker shim install failed",
				[]string{
					err.Error(),
					"",
					"To overwrite an existing file, quit c9s and run:",
					fmt.Sprintf("  c9s install-docker-shim --path %s --force", path),
				},
				modals.InfoError,
				m.palette,
			)
			m.stack.Push(modal)
			return m, modal.Init()
		}
		m.toast = fmt.Sprintf("docker shim installed → %s", path)
		body := []string{
			fmt.Sprintf("Wrote shim to: %s", path),
			fmt.Sprintf("Make sure %s is on your PATH.", filepath.Dir(path)),
			"If your shell caches command lookups, run `hash -r`.",
		}
		level := modals.InfoOK
		if len(others) > 0 {
			level = modals.InfoWarning
			body = append(body, "", "Existing docker binaries detected on PATH:")
			for _, p := range others {
				body = append(body, "  • "+p)
			}
			body = append(body, "Ensure the shim's directory is BEFORE these on PATH or",
				"the existing binary will continue to win.")
		}
		if dockershim.DockerDesktopInstalled() {
			if level == modals.InfoOK {
				level = modals.InfoWarning
			}
			body = append(body, "", "Docker Desktop is installed at /Applications/Docker.app.",
				"Its bin directory may also be on PATH; check `which docker`",
				"after restarting your shell.")
		}
		modal := modals.NewInfo("Docker shim installed", body, level, m.palette)
		m.stack.Push(modal)
		return m, modal.Init()

	case "uninstall-docker-shim":
		path := arg
		if path == "" {
			def, err := dockershim.DefaultPath()
			if err != nil {
				m.toast = fmt.Sprintf("uninstall-docker-shim: %v", err)
				return m, nil
			}
			path = def
		}
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		if err := dockershim.Uninstall(path); err != nil {
			m.toast = "uninstall-docker-shim failed"
			m.logError("dockershim.uninstall", path, fmt.Sprintf("uninstall failed: %v", err), err.Error())
			modal := modals.NewInfo(
				"Docker shim uninstall failed",
				[]string{err.Error()},
				modals.InfoError,
				m.palette,
			)
			m.stack.Push(modal)
			return m, modal.Init()
		}
		m.toast = fmt.Sprintf("docker shim removed → %s", path)
		modal := modals.NewInfo(
			"Docker shim removed",
			[]string{fmt.Sprintf("Removed: %s", path)},
			modals.InfoOK,
			m.palette,
		)
		m.stack.Push(modal)
		return m, modal.Init()

	case "prune":
		// :prune is contextual to the active screen. Today only the
		// containers screen handles ConfirmResultMsg{Tag:"prune"}; on
		// other screens we surface a hint rather than silently no-op.
		if m.active != "containers" {
			m.toast = ":prune currently only works on the containers screen"
			return m, nil
		}
		// Forward a synthetic key event so the screen's existing prune
		// keybinding handler fires the same confirm-modal flow as the
		// Shift+P hotkey. This avoids duplicating the lookup logic here
		// and keeps the screen as the single source of truth for what
		// "prune" means in its context.
		if scr, ok := m.screens["containers"]; ok {
			newScr, cmd := scr.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
			m.screens["containers"] = newScr
			return m, cmd
		}
		return m, nil

	case "create":
		modal := modals.NewTextInput("create", "Resource name to create:", arg, m.palette)
		m.stack.Push(modal)
		return m, modal.Init()
	case "tag":
		modal := modals.NewTextInput("tag", "New tag (source:dest):", arg, m.palette)
		m.stack.Push(modal)
		return m, modal.Init()
	case "save":
		modal := modals.NewTextInput("save", "Save image (ref → tar path):", arg, m.palette)
		m.stack.Push(modal)
		return m, modal.Init()
	case "load":
		modal := modals.NewTextInput("load", "Load image from tar path:", arg, m.palette)
		m.stack.Push(modal)
		return m, modal.Init()
	case "pull":
		if arg == "" {
			m.toast = "usage: :pull <ref>"
			return m, nil
		}
		stream, err := m.client.StreamPull(context.Background(), arg)
		if err != nil {
			m.toast = fmt.Sprintf("pull failed: %v", err)
			m.logError("image.pull", arg, fmt.Sprintf("pull failed: %v", err), err.Error())
			return m, nil
		}
		modal := modals.NewProgressModel(jobs.KindPull, arg, stream, m.clk)
		m.stack.Push(progressModalWrap{p: modal})
		return m, modal.Init()
	case "push":
		if arg == "" {
			m.toast = "usage: :push <ref>"
			return m, nil
		}
		stream, err := m.client.StreamPush(context.Background(), arg)
		if err != nil {
			m.toast = fmt.Sprintf("push failed: %v", err)
			m.logError("image.push", arg, fmt.Sprintf("push failed: %v", err), err.Error())
			return m, nil
		}
		modal := modals.NewProgressModel(jobs.KindPush, arg, stream, m.clk)
		m.stack.Push(progressModalWrap{p: modal})
		return m, modal.Init()

	case "skin":
		if arg == "" {
			m.toast = "usage: :skin <name> · try :skins to list"
			return m, nil
		}
		p, err := theme.LoadSkin(arg)
		if err != nil {
			m.toast = fmt.Sprintf("skin failed: %v", err)
			m.logError("skin.load", arg, fmt.Sprintf("skin failed: %v", err), err.Error())
			return m, nil
		}
		m.palette = p
		// Broadcast palette change to every registered screen so each
		// can update its colors without losing state (filter, marks,
		// sort, scroll, etc.). Screens that don't care about the
		// palette ignore the message.
		for id, scr := range m.screens {
			newScr, _ := scr.Update(screens.PaletteChangedMsg{P: p})
			m.screens[id] = newScr
		}
		m.statusBar = NewStatusBar(p)
		m.skinName = arg
		// Persist to config.toml so this skin loads automatically next time.
		// Surface a persistence failure in the toast so the user sees it
		// immediately rather than silently relying on the error log.
		if err := config.SaveSkin(arg); err != nil {
			m.toast = fmt.Sprintf("loaded skin: %s (but failed to persist: %v)", arg, err)
			m.logError("skin.persist", arg, fmt.Sprintf("could not save skin: %v", err), err.Error())
		} else {
			m.toast = fmt.Sprintf("loaded skin: %s", arg)
		}
		return m, nil

	case "skins":
		// Open an interactive skin picker. Enter on a row applies it.
		skins := theme.ListSkins()
		if len(skins) == 0 {
			m.toast = "no skins found"
			return m, nil
		}
		modal := modals.NewSkinPicker(skins, m.palette)
		m.stack.Push(modal)
		m.crumbs.Push(breadcrumbs.Crumb{Label: "modal:Skins"})
		return m, modal.Init()

	case "commands", "cmds":
		// Show the full palette command catalog.
		var lines []string
		groups := map[string][]palette.Command{}
		for _, c := range palette.Catalog() {
			groups[c.Group] = append(groups[c.Group], c)
		}
		for _, g := range []string{"screen", "system", "k9s", "action", "config", "meta"} {
			cs, ok := groups[g]
			if !ok {
				continue
			}
			lines = append(lines, "## "+strings.ToUpper(g))
			for _, c := range cs {
				usage := ""
				if c.Usage != "" {
					usage = " " + c.Usage
				}
				aliases := ""
				if len(c.Aliases) > 0 {
					aliases = " (" + strings.Join(c.Aliases, ", ") + ")"
				}
				lines = append(lines, fmt.Sprintf("  :%s%s%s — %s", c.Name, usage, aliases, c.Description))
			}
			lines = append(lines, "")
		}
		body := strings.Join(lines, "\n")
		m.stack.Push(modals.NewInspect("Palette commands", []byte(body), m.palette))
		return m, m.stack.Top().Init()

	case "import-skin":
		if arg == "" {
			m.toast = "usage: :import-skin <path/to/k9s-skin.yaml>"
			return m, nil
		}
		outPath, err := theme.ImportK9sSkin(arg)
		if err != nil {
			m.toast = fmt.Sprintf("import-skin failed: %v", err)
			m.logError("skin.import", arg, fmt.Sprintf("import-skin failed: %v", err), err.Error())
			return m, nil
		}
		m.toast = fmt.Sprintf("imported to: %s", outPath)
		return m, nil

	default:
		m.toast = fmt.Sprintf("unknown command: %q", cmd)
		m.logError("command.unknown", cmd, fmt.Sprintf("unknown command: %q", cmd), "")
		return m, nil
	}
}

// logError logs an error to the error log if it's configured.
func (m *Model) logError(op, resource, message, detail string) {
	if m.errorLog != nil {
		_ = m.errorLog.Log(log.Entry{
			Time:     m.clk.Now(),
			Op:       op,
			Resource: resource,
			Message:  message,
			Detail:   detail,
		})
	}
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	if m.showSplash {
		return m.splash.View()
	}

	// Build header
	var header string
	if m.headerVisible {
		scr, ok := m.screens[m.active]
		title := m.active
		if ok {
			title = scr.Title()
		}

		// k9s-style colour roles. Every style explicitly sets the body bg
		// because lipgloss emits ANSI resets between styled spans, and
		// without a per-style bg, terminal-default (often black) leaks
		// through gaps inside columns.
		labelStyle := lipgloss.NewStyle().Foreground(m.palette.Accent2()).Background(m.palette.Bg).Bold(true)
		valueStyle := lipgloss.NewStyle().Foreground(m.palette.Fg).Background(m.palette.Bg).Bold(true)
		keyStyle := lipgloss.NewStyle().Foreground(m.palette.Accent).Background(m.palette.Bg).Bold(true)
		descStyle := lipgloss.NewStyle().Foreground(m.palette.Fg).Background(m.palette.Bg)
		logoStyle := lipgloss.NewStyle().Foreground(m.palette.Accent).Background(m.palette.Bg).Bold(true)

		// Build a short summary that fits in the context column.
		runtimeVer := "—"
		if m.caps.Version != "" {
			runtimeVer = m.caps.Version
		}
		shortSummary := truncateRunes(scrSummary(scr), 24)

		// Logo strings (6 rows).
		logoLines := []string{
			" ██████╗ ██████╗ ███████╗",
			"██╔════╝██╔═══██╗██╔════╝",
			"██║     ╚██████╔╝███████╗",
			"██║      ╚═══██║ ╚════██║",
			"╚██████╗  █████╔╝███████║",
			" ╚═════╝  ╚════╝ ╚══════╝",
		}
		// Per-screen action hotkeys (6 rows).
		actionRows := renderHotkeyRows(scr, keyStyle, descStyle, 6)

		// Constant nav hotkeys (6 rows). Trailing space is inside keyStyle so
		// there is no unstyled gap before the description.
		navKeys := []string{"<:>      ", "</>      ", "<?>      ", "<S>      ", "<space>  ", "<q>      "}
		navDescs := []string{"Palette", "Filter", "Help", "Sort", "Mark", "Quit"}
		navRows := make([]string, 6)
		for i := 0; i < 6; i++ {
			navRows[i] = keyStyle.Render(navKeys[i]) + descStyle.Render(navDescs[i])
		}

		// Context info (6 rows).
		ctxLabels := []string{"Context:  ", "Runtime:  ", "c9s Rev:  ", strings.ToUpper(title) + ":  ", "Skin:     ", "Mode:     "}
		ctxValues := []string{"apple container", runtimeVer, version.Short(), shortSummary, m.skinName, modeName(m.readonly)}
		// Pad ctx labels to consistent prefix width so values line up.
		maxLabel := 0
		for _, l := range ctxLabels {
			if len(l) > maxLabel {
				maxLabel = len(l)
			}
		}
		ctxRows := make([]string, 6)
		for i := 0; i < 6; i++ {
			ctxRows[i] = labelStyle.Render(padRight(ctxLabels[i], maxLabel)) + valueStyle.Render(ctxValues[i])
		}

		// Layout widths.
		const col1W = 38
		const col2W = 22
		const col3W = 22
		const logoW = 28
		spacerW := m.width - col1W - col2W - col3W - logoW - 4
		if spacerW < 0 {
			spacerW = 0
		}

		// Compose 6 full-width header rows. Each row uses a single Render
		// per column with the bg style applied — that way every padding
		// character has the skin's bg, with no unstyled gaps.
		bg := lipgloss.NewStyle().Background(m.palette.Bg).Foreground(m.palette.Fg)
		var rows []string
		for i := 0; i < 6; i++ {
			c1 := bg.Width(col1W).Render(ctxRows[i])
			c2 := bg.Width(col2W).Render(actionRows[i])
			c3 := bg.Width(col3W).Render(navRows[i])
			sp := bg.Width(spacerW).Render("")
			lg := bg.Width(logoW).Align(lipgloss.Right).Render(logoStyle.Render(logoLines[i]))
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, c1, c2, c3, sp, lg))
		}
		// Wrap the 6 banner rows with vertical + horizontal padding so the
		// logo + columns have noticeable breathing room from the top.
		// Padding(top, right, bottom, left) — top=2, bottom=1 to keep
		// the banner snug against the body table (k9s-style).
		titleBar := lipgloss.NewStyle().
			Background(m.palette.Bg).
			Width(m.width).
			Padding(2, 2, 1, 2).
			Render(lipgloss.JoinVertical(lipgloss.Left, rows...))

		breadcrumbBar := ""
		if m.crumbs.Len() > 1 {
			breadcrumbBar = lipgloss.NewStyle().
				Foreground(m.palette.Fg).
				Background(m.palette.Bg).
				Width(m.width).
				Render(" " + m.crumbs.Render(m.width-2))
		}

		if breadcrumbBar != "" {
			header = titleBar + "\n" + breadcrumbBar
		} else {
			header = titleBar
		}
	}

	// Build body
	bodyHeight := m.height - 2 // status bar + palette line
	if m.headerVisible {
		bodyHeight -= 9 // banner: 2 rows top pad + 6 content + 1 bottom pad
		if m.crumbs.Len() > 1 {
			bodyHeight -= 1
		}
	}

	body := ""
	if scr, ok := m.screens[m.active]; ok {
		body = scr.View(m.width, bodyHeight)
	}

	// Render modal overlay if present, centered and surrounded by skin bg.
	var modalOverlay string
	if !m.stack.Empty() {
		modal := m.stack.Top()
		modalContent := modal.View(m.width, bodyHeight)
		modalOverlay = lipgloss.Place(
			m.width, bodyHeight,
			lipgloss.Center, lipgloss.Center,
			modalContent,
			lipgloss.WithWhitespaceBackground(m.palette.Bg),
			lipgloss.WithWhitespaceForeground(m.palette.Bg),
		)
	}

	// Build status bar
	summary := ""
	if scr, ok := m.screens[m.active]; ok {
		summary = scr.Summary()
	}
	sb := m.statusBar.Update(StatusUpdate{
		Screen:  m.active,
		Summary: summary,
		Hint:    m.hint(),
		Toast:   m.toast,
	}).View(m.width, m.readonly)

	// Build palette line + autocomplete dropdown
	var paletteLine string
	var paletteDropdown string
	if m.cmdActive {
		bgFill := lipgloss.NewStyle().Background(m.palette.Bg).Foreground(m.palette.Fg)
		paletteLine = bgFill.Width(m.width).Render(
			lipgloss.NewStyle().Foreground(m.palette.Accent).Background(m.palette.Bg).Render(":" + m.cmdBuf + "▏"),
		)
		// Show matching commands as a dropdown above the prompt. All cells
		// have explicit bg so the dropdown paints with the skin colour.
		matches := palette.Match(m.cmdBuf, palette.Catalog())
		if len(matches) > 8 {
			matches = matches[:8]
		}
		if len(matches) > 0 {
			dim := lipgloss.NewStyle().Foreground(m.palette.Accent2()).Background(m.palette.Bg)
			fg := lipgloss.NewStyle().Foreground(m.palette.Fg).Background(m.palette.Bg)
			accent := lipgloss.NewStyle().Foreground(m.palette.Accent).Background(m.palette.Bg).Bold(true)
			header := bgFill.Width(m.width).Render(
				dim.Render(fmt.Sprintf(" %d match%s · Tab to complete · Enter to run", len(matches), pluralize(len(matches)))),
			)
			rows := []string{header}
			for _, c := range matches {
				name := accent.Render(c.Name)
				rest := ""
				if c.Usage != "" {
					rest += dim.Render(" " + c.Usage)
				}
				if len(c.Aliases) > 0 {
					rest += dim.Render(" (" + strings.Join(c.Aliases, ", ") + ")")
				}
				rest += fg.Render(" — " + c.Description)
				row := bgFill.Width(m.width).Render(fg.Render(" ") + name + rest)
				rows = append(rows, row)
			}
			paletteDropdown = strings.Join(rows, "\n")
		}
	}

	// Combine
	var parts []string
	if m.headerVisible {
		parts = append(parts, header)
	}

	// When palette is open and there's a dropdown, reduce body height to fit it.
	dropdownHeight := 0
	if paletteDropdown != "" {
		dropdownHeight = strings.Count(paletteDropdown, "\n") + 1
	}

	if modalOverlay != "" {
		parts = append(parts, modalOverlay)
	} else {
		bodyRendered := lipgloss.NewStyle().
			Width(m.width).
			Height(bodyHeight - dropdownHeight).
			Foreground(m.palette.Fg).
			Background(m.palette.Bg).
			Render(body)
		parts = append(parts, bodyRendered)
	}
	if paletteDropdown != "" {
		parts = append(parts, paletteDropdown)
	}
	parts = append(parts, paletteLine, sb)

	// Wrap the entire app in the skin's bg/fg so empty cells in light themes
	// don't fall through to the terminal's default (usually black) bg.
	out := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.NewStyle().
		Foreground(m.palette.Fg).
		Background(m.palette.Bg).
		Render(out)
}

func pluralize(n int) string {
	if n == 1 {
		return ""
	}
	return "es"
}

// numberShortcut returns the screen name for a number key, or "" if not a shortcut.
// Mirrors k9s-style numeric navigation.
func numberShortcut(key string) string {
	switch key {
	case "1":
		return "containers"
	case "2":
		return "images"
	case "3":
		return "volumes"
	case "4":
		return "networks"
	case "5":
		return "builder"
	case "6":
		return "registry"
	case "7":
		return "system"
	case "8":
		return "pulses"
	case "9":
		return "xray"
	case "0":
		return "pinned"
	}
	return ""
}

// SetSkinName updates the skin name displayed in the header. Called by main
// after loading a persisted skin so the banner doesn't show 'default' when
// the user has actually chosen a different theme.
func (m *Model) SetSkinName(name string) {
	if name != "" {
		m.skinName = name
	}
}

// scrSummary returns the screen's Summary string, or "—" if nil.
func scrSummary(s screens.Screen) string {
	if s == nil {
		return "—"
	}
	if sum := s.Summary(); sum != "" {
		return sum
	}
	return "—"
}

// modeName returns "read-only" or "read-write" for the readonly flag.
func modeName(readonly bool) string {
	if readonly {
		return "read-only"
	}
	return "read-write"
}

// renderHotkeyRows builds up to `n` hotkey lines from the active screen's
// keymap. Per-screen actions (logs/inspect/shell/run/etc.) are surfaced;
// global navigation bindings shown in col3 are skipped.
func renderHotkeyRows(scr screens.Screen, keyStyle, descStyle lipgloss.Style, n int) []string {
	if scr == nil {
		out := make([]string, n)
		return out
	}
	// Bindings the global header column 3 already advertises — skip these.
	skip := map[string]bool{
		"palette": true, "filter": true, "help": true, "sort": true,
		"mark": true, "mark_all": true, "quit": true, "interrupt": true,
		"escape": true, "refresh": true, "header_toggle": true,
		"up": true, "down": true, "top": true, "bottom": true,
		"left": true, "right": true,
	}

	km := scr.Hotkeys()
	if km == nil {
		out := make([]string, n)
		return out
	}

	// Preferred order so the most useful actions show first.
	preferred := []string{"logs", "inspect", "shell", "run", "tag", "push", "pull", "stop", "restart", "kill", "delete", "create", "prune", "set_default", "pin", "pause"}
	rendered := make([]string, 0, n)
	used := map[string]bool{}

	addRow := func(name string, b keymap.Binding) {
		if len(rendered) >= n {
			return
		}
		key := primaryKey(b.Keys)
		if key == "" {
			return
		}
		help := b.Help
		if help == "" {
			help = name
		}
		// Build the row as a continuous styled string. Inserting a literal
		// space between two Render() calls would create an unstyled gap
		// because lipgloss emits \x1b[0m resets between spans. Putting the
		// trailing space inside the keyStyle Render keeps the gap painted.
		keyText := padRight(fmt.Sprintf("<%s>", key), 9)
		rendered = append(rendered, keyStyle.Render(keyText)+descStyle.Render(help))
		used[name] = true
	}

	for _, name := range preferred {
		if skip[name] {
			continue
		}
		if b, ok := km.Get(name); ok {
			addRow(name, b)
		}
	}
	if len(rendered) < n {
		for _, name := range km.Names() {
			if used[name] || skip[name] {
				continue
			}
			if b, ok := km.Get(name); ok {
				addRow(name, b)
			}
			if len(rendered) >= n {
				break
			}
		}
	}
	for len(rendered) < n {
		rendered = append(rendered, "")
	}
	return rendered
}

func primaryKey(keys []string) string {
	if len(keys) == 0 {
		return ""
	}
	return keys[0]
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

// truncateRunes truncates a string to at most n visible runes.
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}

func (m Model) hint() string {
	return ": command  ?  help  q  quit"
}
