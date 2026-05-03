package palette

// Command describes a palette command for autocomplete and discovery.
type Command struct {
	// Name is the canonical command name (typed after `:`).
	Name string
	// Aliases are alternative names that resolve to this command.
	Aliases []string
	// Description is a short help blurb shown in the autocomplete dropdown.
	Description string
	// Usage shows arguments after the command, e.g. "<image>".
	Usage string
	// Group tags the command for grouping in help (e.g., "screen", "lifecycle").
	Group string
}

// Catalog returns the canonical list of palette commands.
func Catalog() []Command {
	return []Command{
		// Navigation — resource screens
		{Name: "containers", Aliases: []string{"c"}, Description: "Switch to containers screen", Group: "screen"},
		{Name: "images", Aliases: []string{"i"}, Description: "Switch to images screen", Group: "screen"},
		{Name: "volumes", Aliases: []string{"v"}, Description: "Switch to volumes screen", Group: "screen"},
		{Name: "networks", Aliases: []string{"n"}, Description: "Switch to networks screen", Group: "screen"},
		{Name: "builder", Aliases: []string{"b"}, Description: "Switch to builder screen", Group: "screen"},
		{Name: "registry", Aliases: []string{"reg"}, Description: "Switch to registry screen", Group: "screen"},
		{Name: "system", Aliases: []string{"sys"}, Description: "Switch to system services", Group: "screen"},

		// System sub-screens
		{Name: "df", Description: "Disk usage breakdown", Group: "system"},
		{Name: "dns", Description: "DNS domains", Group: "system"},
		{Name: "property", Description: "System properties", Group: "system"},
		{Name: "kernel", Description: "Kernel configuration", Group: "system"},
		{Name: "logs", Description: "System logs viewer", Group: "system"},

		// k9s parity screens
		{Name: "pulses", Description: "Live health dashboard", Group: "k9s"},
		{Name: "xray", Description: "Resource relationships tree", Group: "k9s"},
		{Name: "jobs", Description: "Background streaming jobs", Group: "k9s"},
		{Name: "pinned", Description: "Pinned resources", Group: "k9s"},
		{Name: "errors", Description: "Error log viewer", Group: "k9s"},

		// Action commands
		{Name: "run", Usage: "<image>", Description: "Launch a new container", Group: "action"},
		{Name: "build", Usage: "<path>", Description: "Build an image from path", Group: "action"},
		{Name: "pull", Usage: "<ref>", Description: "Pull image from registry", Group: "action"},
		{Name: "push", Usage: "<ref>", Description: "Push image to registry", Group: "action"},
		{Name: "create", Usage: "<name>", Description: "Create resource on current screen", Group: "action"},
		{Name: "tag", Usage: "<src> <dst>", Description: "Tag an image", Group: "action"},
		{Name: "save", Usage: "<ref> <tar>", Description: "Save image to tar file", Group: "action"},
		{Name: "load", Usage: "<tar>", Description: "Load image from tar file", Group: "action"},
		{Name: "login", Usage: "<host>", Description: "Login to a registry", Group: "action"},

		// Customization
		{Name: "skin", Usage: "<name>", Description: "Switch to a TOML skin (try :skins to list)", Group: "config"},
		{Name: "skins", Description: "List available bundled skins", Group: "config"},
		{Name: "import-skin", Usage: "<path>", Description: "Import a k9s YAML skin", Group: "config"},

		// Meta
		{Name: "help", Aliases: []string{"?"}, Description: "Show help overlay for the active screen", Group: "meta"},
		{Name: "quit", Aliases: []string{"q", "exit"}, Description: "Quit c9s", Group: "meta"},
	}
}

// Match returns the catalog entries whose Name or Alias starts with prefix.
// Results are ordered: exact name matches first, then prefix matches by length.
func Match(prefix string, catalog []Command) []Command {
	if prefix == "" {
		return catalog
	}
	var exact, name, alias []Command
	for _, c := range catalog {
		if c.Name == prefix {
			exact = append(exact, c)
			continue
		}
		matched := false
		if hasPrefix(c.Name, prefix) {
			name = append(name, c)
			matched = true
		}
		if !matched {
			for _, a := range c.Aliases {
				if hasPrefix(a, prefix) {
					alias = append(alias, c)
					break
				}
			}
		}
	}
	out := make([]Command, 0, len(exact)+len(name)+len(alias))
	out = append(out, exact...)
	out = append(out, name...)
	out = append(out, alias...)
	return out
}

func hasPrefix(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		a, b := s[i], prefix[i]
		if a >= 'A' && a <= 'Z' {
			a += 32
		}
		if b >= 'A' && b <= 'Z' {
			b += 32
		}
		if a != b {
			return false
		}
	}
	return true
}
