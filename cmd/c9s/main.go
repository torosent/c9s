// Command c9s is the entrypoint for the TUI.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/cli/demodata"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/config"
	"github.com/torosent/c9s/internal/plugins"
	"github.com/torosent/c9s/internal/ui"
	"github.com/torosent/c9s/internal/ui/theme"
	"github.com/torosent/c9s/internal/version"
)

func main() {
	// Check for subcommands first
	if len(os.Args) > 1 && os.Args[1] == "import-skin" {
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: c9s import-skin <k9s-skin.yaml>")
			os.Exit(1)
		}
		outPath, err := theme.ImportK9sSkin(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "import-skin failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Imported k9s skin to: %s\n", outPath)
		return
	}

	showVersion := flag.Bool("version", false, "print version and exit")
	readonly := flag.Bool("readonly", false, "enable read-only mode (disables destructive operations)")
	demoData := flag.Bool("demo-data", false, "run with populated demo data instead of real container runtime")
	flag.Parse()
	if *showVersion {
		fmt.Println(version.String())
		return
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "c9s panicked: %v\n%s\n", r, debug.Stack())
			os.Exit(2)
		}
	}()

	// Load configuration
	cfg, err := config.LoadFromXDG()
	if err != nil {
		fmt.Fprintf(os.Stderr, "c9s: config load failed: %v\n", err)
		// Continue with defaults
		cfg = config.Default()
	}

	// Command-line flag overrides config
	if *readonly {
		cfg.UI.ReadOnly = true
	}

	// Load plugins
	pluginList, err := plugins.LoadFromXDG()
	if err != nil {
		fmt.Fprintf(os.Stderr, "c9s: plugin load failed: %v\n", err)
		// Continue without plugins
		pluginList = nil
	}

	// Build client: demo data or real runtime
	var client cli.Client
	if *demoData {
		client = demodata.NewFake(clock.Real())
	} else {
		client = cli.NewDefaultClient()
	}

	// Load the persisted skin if the user previously selected one. Falls
	// back to default dark on any error so a corrupted skin doesn't break
	// startup.
	palette := theme.DefaultDark()
	skinName := "default"
	if cfg.Theme.Name != "" && cfg.Theme.Name != "dark" {
		if loaded, err := theme.LoadSkin(cfg.Theme.Name); err == nil {
			palette = loaded
			skinName = cfg.Theme.Name
		}
	}

	app := ui.NewApp(client, clock.Real(), palette, cfg, pluginList)
	app.SetSkinName(skinName)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "c9s:", err)
		os.Exit(1)
	}
}
