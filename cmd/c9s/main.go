// Command c9s is the entrypoint for the TUI.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	tea "charm.land/bubbletea/v2"

	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/cli/demodata"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/config"
	"github.com/torosent/c9s/internal/dockershim"
	"github.com/torosent/c9s/internal/ui"
	"github.com/torosent/c9s/internal/ui/blockingwriter"
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

	if len(os.Args) > 1 && os.Args[1] == "install-docker-shim" {
		runInstallDockerShim(os.Args[2:])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "uninstall-docker-shim" {
		runUninstallDockerShim(os.Args[2:])
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

	app := ui.NewApp(client, clock.Real(), palette, cfg)
	app.SetSkinName(skinName)
	// Wrap stdout in a blocking writer so the bubbletea v2 renderer can
	// always flush a full frame. After tea.ExecProcess restores the
	// terminal on macOS, stdout can be left in non-blocking mode and
	// large frames (~10 KB) hit EAGAIN at the kernel TTY buffer (~1 KB),
	// which the renderer treats as fatal — leaving the screen stuck on
	// a partially drawn frame after the user exits an exec'd shell.
	p := tea.NewProgram(app, tea.WithOutput(blockingwriter.New(os.Stdout)))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "c9s:", err)
		os.Exit(1)
	}
}

func runInstallDockerShim(args []string) {
	fs := flag.NewFlagSet("install-docker-shim", flag.ExitOnError)
	path := fs.String("path", "", "where to write the shim (default: ~/.local/bin/docker)")
	force := fs.Bool("force", false, "overwrite an existing file at the target path")
	_ = fs.Parse(args)

	target := *path
	if target == "" {
		def, err := dockershim.DefaultPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "install-docker-shim: %v\n", err)
			os.Exit(1)
		}
		target = def
	}
	// Resolve to absolute before reporting; relative --path values
	// otherwise produce a misleading PATH-search hint.
	if abs, err := filepath.Abs(target); err == nil {
		target = abs
	}

	// Detect any existing docker so the user knows what the shim is
	// going to compete with on PATH and which app might be running.
	others, _ := dockershim.DetectExistingDocker(target)
	if len(others) > 0 {
		fmt.Fprintln(os.Stderr, "Warning: an existing docker binary is already on PATH:")
		for _, p := range others {
			fmt.Fprintf(os.Stderr, "  - %s\n", p)
		}
		fmt.Fprintln(os.Stderr, "After install, make sure the shim's directory is BEFORE")
		fmt.Fprintln(os.Stderr, "the directory containing the existing docker on PATH, or")
		fmt.Fprintln(os.Stderr, "the existing binary will continue to win.")
		fmt.Fprintln(os.Stderr)
	}
	if dockershim.DockerDesktopInstalled() {
		fmt.Fprintln(os.Stderr, "Warning: Docker Desktop is installed at /Applications/Docker.app.")
		fmt.Fprintln(os.Stderr, "While Docker Desktop is running, the system docker daemon will")
		fmt.Fprintln(os.Stderr, "still respond to whichever 'docker' binary your shell resolves")
		fmt.Fprintln(os.Stderr, "first. The shim only redirects CLI calls to Apple containers; it")
		fmt.Fprintln(os.Stderr, "doesn't affect Docker Desktop's running containers either way.")
		fmt.Fprintln(os.Stderr)
	}

	if err := dockershim.Install(target, *force); err != nil {
		fmt.Fprintf(os.Stderr, "install-docker-shim: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Installed c9s docker shim to: %s\n", target)
	fmt.Println()
	fmt.Println("Make sure the install directory is on PATH and shadows any real docker:")
	fmt.Printf("  echo $PATH | tr ':' '\\n' | grep -F %q\n", filepath.Dir(target))
	fmt.Println("If your shell caches command lookups, run `hash -r` (bash/zsh).")
}

func runUninstallDockerShim(args []string) {
	fs := flag.NewFlagSet("uninstall-docker-shim", flag.ExitOnError)
	path := fs.String("path", "", "shim path to remove (default: ~/.local/bin/docker)")
	_ = fs.Parse(args)

	target := *path
	if target == "" {
		def, err := dockershim.DefaultPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "uninstall-docker-shim: %v\n", err)
			os.Exit(1)
		}
		target = def
	}
	if abs, err := filepath.Abs(target); err == nil {
		target = abs
	}

	if err := dockershim.Uninstall(target); err != nil {
		fmt.Fprintf(os.Stderr, "uninstall-docker-shim: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Removed c9s docker shim at: %s\n", target)
}
