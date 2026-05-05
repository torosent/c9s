package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/config"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/theme"
)

func newTestApp(t *testing.T) Model {
	t.Helper()
	fake := cli.NewFake()
	app := NewApp(fake, clock.NewFake(time.Unix(0, 0)), theme.DefaultDark(), config.Default())
	// Force size + dismiss splash
	m, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app = m.(Model)
	app.showSplash = false
	return app
}

func TestRunCommandSwitchScreens(t *testing.T) {
	cases := map[string]string{
		"images":     "images",
		"i":          "images",
		"volumes":    "volumes",
		"v":          "volumes",
		"networks":   "networks",
		"n":          "networks",
		"builder":    "builder",
		"b":          "builder",
		"registry":   "registry",
		"reg":        "registry",
		"system":     "system",
		"sys":        "system",
		"df":         "df",
		"dns":        "dns",
		"property":   "property",
		"kernel":     "kernel",
		"logs":       "syslogs",
		"c":          "containers",
		"containers": "containers",
	}
	for cmd, want := range cases {
		app := newTestApp(t)
		updated, _ := app.runCommand(cmd)
		got := updated.(Model).active
		if got != want {
			t.Errorf("runCommand(%q) = active %q, want %q", cmd, got, want)
		}
	}
}

func TestRunCommandUnknownProducesToast(t *testing.T) {
	app := newTestApp(t)
	updated, _ := app.runCommand("nonexistent")
	got := updated.(Model).toast
	if !strings.Contains(got, "unknown") {
		t.Errorf("expected unknown toast, got %q", got)
	}
}

func TestRunCommandQuitReturnsTeaQuit(t *testing.T) {
	app := newTestApp(t)
	_, cmd := app.runCommand("q")
	if cmd == nil {
		t.Fatal("expected tea.Quit cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("got %T, want tea.QuitMsg", msg)
	}
}

func TestRunCommandHelpOpensModal(t *testing.T) {
	app := newTestApp(t)
	updated, _ := app.runCommand("help")
	res := updated.(Model)
	if res.stack.Empty() {
		t.Error("expected modal pushed")
	}
}

func TestRunCommandRunOpensRunForm(t *testing.T) {
	app := newTestApp(t)
	updated, _ := app.runCommand("run alpine")
	res := updated.(Model)
	if res.stack.Empty() {
		t.Fatal("expected run form modal")
	}
	top := res.stack.Top()
	if _, ok := top.(modals.RunFormModel); !ok {
		t.Errorf("expected RunFormModel, got %T", top)
	}
}

func TestRunCommandBuildOpensBuildForm(t *testing.T) {
	app := newTestApp(t)
	updated, _ := app.runCommand("build ./api")
	res := updated.(Model)
	top := res.stack.Top()
	if _, ok := top.(modals.BuildFormModel); !ok {
		t.Errorf("expected BuildFormModel, got %T", top)
	}
}

func TestRunCommandLoginOpensLogin(t *testing.T) {
	app := newTestApp(t)
	updated, _ := app.runCommand("login ghcr.io")
	res := updated.(Model)
	top := res.stack.Top()
	if _, ok := top.(modals.LoginModel); !ok {
		t.Errorf("expected LoginModel, got %T", top)
	}
}

func TestRunCommandTextInputCommands(t *testing.T) {
	for _, c := range []string{"create foo", "tag src", "save ref out.tar", "load /tmp/in.tar"} {
		app := newTestApp(t)
		updated, _ := app.runCommand(c)
		res := updated.(Model)
		top := res.stack.Top()
		if _, ok := top.(modals.TextInputModel); !ok {
			t.Errorf("%q: expected TextInputModel, got %T", c, top)
		}
	}
}

func TestRunCommandPullPushNoArgs(t *testing.T) {
	app := newTestApp(t)
	updated, _ := app.runCommand("pull")
	got := updated.(Model).toast
	if !strings.Contains(got, "usage") {
		t.Errorf("expected usage toast, got %q", got)
	}
	app2 := newTestApp(t)
	updated, _ = app2.runCommand("push")
	got = updated.(Model).toast
	if !strings.Contains(got, "usage") {
		t.Errorf("expected usage toast, got %q", got)
	}
}

func TestRunCommandPullOpensProgress(t *testing.T) {
	app := newTestApp(t)
	updated, cmd := app.runCommand("pull alpine")
	if cmd == nil {
		t.Fatal("expected init cmd from pull")
	}
	res := updated.(Model)
	if res.stack.Empty() {
		t.Error("expected progress modal pushed")
	}
}

func TestUpdateRouteRunSubmitted(t *testing.T) {
	app := newTestApp(t)
	updated, cmd := app.Update(modals.RunSubmittedMsg{Opts: cli.RunOpts{Image: "alpine"}})
	if cmd == nil {
		t.Fatal("expected progress init cmd")
	}
	res := updated.(Model)
	if res.stack.Empty() {
		t.Error("expected progress modal pushed")
	}
}

func TestUpdateRouteBuildSubmitted(t *testing.T) {
	app := newTestApp(t)
	updated, cmd := app.Update(modals.BuildSubmittedMsg{Opts: cli.BuildOpts{Tag: "x"}})
	if cmd == nil {
		t.Fatal("expected progress init cmd")
	}
	res := updated.(Model)
	if res.stack.Empty() {
		t.Error("expected progress modal pushed")
	}
}

func TestUpdateRouteCancelMessages(t *testing.T) {
	for _, msg := range []tea.Msg{modals.RunCancelledMsg{}, modals.BuildCancelledMsg{}} {
		app := newTestApp(t)
		updated, _ := app.Update(msg)
		if updated.(Model).toast != "cancelled" {
			t.Errorf("expected 'cancelled' toast for %T, got %q", msg, updated.(Model).toast)
		}
	}
}

func TestUpdateRouteStatusMsg(t *testing.T) {
	app := newTestApp(t)
	updated, _ := app.Update(screens.StatusMsg{Toast: "hello"})
	if updated.(Model).toast != "hello" {
		t.Errorf("expected toast 'hello', got %q", updated.(Model).toast)
	}
}

func TestAliasResolution(t *testing.T) {
	app := newTestApp(t)
	// Add some aliases to the app model
	app.aliases = map[string]string{
		"kpods": "containers",
		"kimg":  "images",
		"kvol":  "volumes",
	}

	cases := map[string]string{
		"kpods": "containers",
		"kimg":  "images",
		"kvol":  "volumes",
	}

	for alias, expectedScreen := range cases {
		updated, _ := app.runCommand(alias)
		got := updated.(Model).active
		if got != expectedScreen {
			t.Errorf("runCommand(%q) with alias = active %q, want %q", alias, got, expectedScreen)
		}
	}
}

func TestReadonlyMode_BlocksDestructiveCommands(t *testing.T) {
	app := newTestApp(t)
	app.readonly = true

	destructive := []string{"delete", "prune", "stop", "kill", "pause", "rm", "rmi"}

	for _, cmd := range destructive {
		updated, _ := app.runCommand(cmd)
		got := updated.(Model).toast
		if !strings.Contains(strings.ToLower(got), "read-only") {
			t.Errorf("command %q in readonly mode: expected 'read-only' toast, got %q", cmd, got)
		}
	}
}

func TestReadonlyMode_AllowsSafeCommands(t *testing.T) {
	app := newTestApp(t)
	app.readonly = true

	// These commands should still work in readonly mode
	safe := []string{"containers", "images", "help", "refresh"}

	for _, cmd := range safe {
		updated, _ := app.runCommand(cmd)
		toast := updated.(Model).toast
		if strings.Contains(strings.ToLower(toast), "read-only") {
			t.Errorf("command %q should work in readonly mode, got toast: %q", cmd, toast)
		}
	}
}

func TestRunCommandSkin(t *testing.T) {
	app := newTestApp(t)

	// Test bundled skins
	for _, skin := range []string{"dark", "light", "k9s-dark", "k9s-light"} {
		updated, _ := app.runCommand("skin " + skin)
		m := updated.(Model)
		if !strings.Contains(m.toast, "loaded skin") {
			t.Errorf("skin %q failed, toast: %q", skin, m.toast)
		}
	}

	// Test nonexistent skin
	updated, _ := app.runCommand("skin nonexistent")
	m := updated.(Model)
	if !strings.Contains(m.toast, "failed") {
		t.Errorf("expected error for nonexistent skin, got toast: %q", m.toast)
	}

	// Test missing arg
	updated, _ = app.runCommand("skin")
	m = updated.(Model)
	if !strings.Contains(m.toast, "usage") {
		t.Errorf("expected usage message, got toast: %q", m.toast)
	}
}

func TestRunCommandImportSkin(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	app := newTestApp(t)

	// Create a sample k9s YAML skin
	k9sYAML := `
k9s:
  fgColor: "#ffffff"
  bgColor: "#000000"
  borderColor: "#333333"
  focusColor: "#00ff00"
  crumbColor: "#666666"
  statusSuccessColor: "#0f0"
  statusWarningColor: "#ff0"
  statusInfoColor: "#00f"
  statusErrorColor: "#f00"
  titleColor: "#eee"
`

	yamlPath := filepath.Join(tmpDir, "test-k9s.yaml")
	if err := os.WriteFile(yamlPath, []byte(k9sYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	// Test import
	updated, _ := app.runCommand("import-skin " + yamlPath)
	m := updated.(Model)
	if !strings.Contains(m.toast, "imported") {
		t.Errorf("expected success toast, got: %q", m.toast)
	}

	// Test missing arg
	updated, _ = app.runCommand("import-skin")
	m = updated.(Model)
	if !strings.Contains(m.toast, "usage") {
		t.Errorf("expected usage message, got: %q", m.toast)
	}

	// Test invalid file
	updated, _ = app.runCommand("import-skin /nonexistent.yaml")
	m = updated.(Model)
	if !strings.Contains(m.toast, "failed") {
		t.Errorf("expected error for nonexistent file, got: %q", m.toast)
	}
}
