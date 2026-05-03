// Command gen-hotkeys generates docs/hotkeys.md from screen implementations.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/jobs"
	"github.com/torosent/c9s/internal/pinned"
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
)

func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func main() {
	fake := cli.NewFake()
	clk := clock.Real()
	p := theme.DefaultDark()
	pinStore, _ := pinned.Load("/tmp/pinned.toml")
	if pinStore == nil {
		pinStore = &pinned.Store{}
	}
	jobMgr := jobs.New(clk)

	screenMap := map[string]screens.Screen{
		"containers": containerscreen.New(fake, clk, p),
		"images":     imagesscreen.New(fake, clk, p),
		"volumes":    volumesscreen.New(fake, clk, p),
		"networks":   networksscreen.New(fake, clk, p),
		"builder":    builderscreen.New(fake, clk, p),
		"registry":   registryscreen.New(fake, clk, p),
		"system":     systemscreen.NewServices(fake, clk, p),
		"df":         systemscreen.NewDF(fake, clk, p),
		"dns":        systemscreen.NewDNS(fake, clk, p),
		"property":   systemscreen.NewProperty(fake, clk, p),
		"kernel":     systemscreen.NewKernel(fake, clk, p),
		"syslogs":    systemscreen.NewLogs(fake, clk, p),
		"errors":     errorsscreen.New("/tmp", clk, p),
		"pinned":     pinnedscreen.New(pinStore, p),
		"xray":       xrayscreen.New(fake, p),
		"pulses":     pulsesscreen.New(fake, clk, p),
		"jobs":       jobsscreen.New(jobMgr, clk, p),
	}

	// Sort screen names for stable output
	var names []string
	for name := range screenMap {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Println("# Hotkeys Reference")
	fmt.Println()
	fmt.Println("This document lists all available hotkeys for each screen in c9s.")
	fmt.Println()
	fmt.Println("## Global Hotkeys")
	fmt.Println()
	fmt.Println("These hotkeys are available on all screens:")
	fmt.Println()
	fmt.Println("| Key | Action |")
	fmt.Println("|-----|--------|")
	fmt.Println("| `q` | Quit |")
	fmt.Println("| `?` | Help |")
	fmt.Println("| `:` | Command palette |")
	fmt.Println("| `/` | Filter |")
	fmt.Println("| `Esc` | Cancel/Close |")
	fmt.Println()

	for _, name := range names {
		screen := screenMap[name]
		km := screen.Hotkeys()
		if km == nil {
			continue
		}

		// Get all binding names and sort them
		bindingNames := km.Names()
		if len(bindingNames) == 0 {
			continue
		}

		fmt.Printf("## Screen: %s\n\n", titleCase(name))
		fmt.Println("| Key | Action | Description |")
		fmt.Println("|-----|--------|-------------|")

		for _, bname := range bindingNames {
			binding, ok := km.Get(bname)
			if !ok {
				continue
			}

			keysStr := strings.Join(binding.Keys, ", ")
			if keysStr == "" {
				continue
			}

			// Format keys with backticks
			var formattedKeys []string
			for _, k := range binding.Keys {
				formattedKeys = append(formattedKeys, fmt.Sprintf("`%s`", k))
			}
			keysDisplay := strings.Join(formattedKeys, ", ")

			help := binding.Help
			if help == "" {
				help = bname
			}

			desc := binding.Description
			if desc == "" {
				desc = help
			}

			fmt.Printf("| %s | %s | %s |\n", keysDisplay, help, desc)
		}
		fmt.Println()
	}

	fmt.Println("## Notes")
	fmt.Println()
	fmt.Println("- Use arrow keys or `j`/`k` to navigate lists")
	fmt.Println("- `Space` to mark/unmark items for batch operations")
	fmt.Println("- `Enter` to select or open details")
	fmt.Println("- Mouse support is available for clicking and scrolling")

	os.Exit(0)
}
