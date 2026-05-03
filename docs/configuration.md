# Configuration

c9s supports a layered configuration system that allows customization of behavior, appearance, and keybindings.

## Configuration Files

Configuration files are loaded from `~/.config/c9s/` (or `$XDG_CONFIG_HOME/c9s/` if set).

### config.toml

Main configuration file with general settings.

**Location:** `~/.config/c9s/config.toml`

**Example:**

```toml
[ui]
readonly = false
header_visible_default = true
refresh_interval_seconds = 3

[stream]
log_buffer_lines = 10000
log_follow_default = true
build_save_logs = false
build_autoclose = true

[theme]
name = "k9s-gruvbox-dark"
```

#### UI Section

- `readonly` (bool): Enable read-only mode, disabling all destructive operations. Default: `false`
- `header_visible_default` (bool): Show table headers by default. Default: `true`
- `refresh_interval_seconds` (int): Auto-refresh interval in seconds. Default: `3`

#### Stream Section

- `log_buffer_lines` (int): Maximum lines to buffer in log viewers. Default: `10000`
- `log_follow_default` (bool): Auto-follow logs by default (tail -f behavior). Default: `true`
- `build_save_logs` (bool): Save build logs to disk. Default: `false`
- `build_autoclose` (bool): Auto-close progress modal on successful build. Default: `true`

#### Theme Section

- `name` (string): Theme name to load. Default: `"k9s-gruvbox-dark"`. Other bundled options include `"dark"`, `"light"`, `"k9s-dark"`, `"k9s-light"`, `"k9s-dracula"`, `"k9s-gruvbox-light"`, `"k9s-monokai"`, `"k9s-nightfox"`, `"k9s-nord"`, `"k9s-one-dark"`, `"k9s-one-light"`, `"k9s-kanagawa"`. See `docs/skins.md`.
- `overrides` (map): Per-key color overrides (see Skins documentation)

### hotkeys.toml

Override default key bindings.

**Location:** `~/.config/c9s/hotkeys.toml`

**Example:**

```toml
[hotkeys]
"quit" = "Q"
"containers.delete" = "x"
"images.delete" = "shift+x"
"filter" = "ctrl+f"
```

Format: `"action" = "key"` where `action` is the binding name and `key` is the new key string.

Supported key formats:
- Single characters: `"a"`, `"Q"` (case-sensitive)
- Special keys: `"esc"`, `"enter"`, `"space"`, `"tab"`, `"up"`, `"down"`, `"left"`, `"right"`
- Ctrl combinations: `"ctrl+c"`, `"ctrl+e"`, `"ctrl+d"`
- Shift combinations: `"shift+k"`, `"shift+g"`

### aliases.toml

Create command aliases for the palette (`:` command mode).

**Location:** `~/.config/c9s/aliases.toml`

**Example:**

```toml
[aliases]
kpods = "containers"
kimg = "images"
kvol = "volumes"
knet = "networks"
ll = "logs"
```

After defining these aliases, you can use `:kpods` instead of `:containers`, etc.

### views.toml

Customize per-screen column visibility, order, and width.

**Location:** `~/.config/c9s/views.toml`

**Example:**

```toml
[containers.columns]
short_id = { visible = true, width = 12, order = 0 }
image = { visible = true, width = 30, order = 1 }
state = { visible = true, width = 10, order = 2 }
uptime = { visible = false }  # hidden
cpu = { visible = true, width = 6, order = 3 }
mem = { visible = true, width = 8, order = 4 }

[images.columns]
repository = { visible = true, width = 40 }
tag = { visible = true, width = 20 }
size = { visible = true, width = 10 }
```

- `visible`: Show or hide the column (default: true)
- `width`: Column width in characters (default: auto)
- `order`: Display order (lower numbers first)

Columns without explicit configuration use defaults.

## Live Reload

c9s watches configuration files for changes and reloads them automatically. Changes take effect immediately without restarting the application.

## Read-Only Mode

Enable read-only mode to prevent accidental modifications:

**Via CLI flag:**
```bash
c9s --readonly
```

**Via config.toml:**
```toml
[ui]
readonly = true
```

In read-only mode:
- Destructive commands are blocked (`delete`, `prune`, `stop`, `kill`, etc.)
- `[READONLY]` indicator appears in the status bar
- All view and navigation commands work normally

## Command-Line Flags

- `--version`: Print version and exit
- `--readonly`: Enable read-only mode (overrides config file)

## Configuration Precedence

1. Command-line flags (highest priority)
2. `~/.config/c9s/config.toml`
3. Built-in defaults (lowest priority)

## Examples

### Minimal Config

```toml
# ~/.config/c9s/config.toml
[ui]
readonly = true
refresh_interval_seconds = 5
```

### Power User Config

```toml
# ~/.config/c9s/config.toml
[ui]
readonly = false
header_visible_default = true
refresh_interval_seconds = 2

[stream]
log_buffer_lines = 50000
log_follow_default = true
build_save_logs = true

[theme]
name = "k9s-gruvbox-dark"
```

```toml
# ~/.config/c9s/hotkeys.toml
[hotkeys]
"quit" = "Q"
"filter" = "f"
"mark" = "m"
```

```toml
# ~/.config/c9s/aliases.toml
[aliases]
c = "containers"
i = "images"
v = "volumes"
n = "networks"
```

## Troubleshooting

### Config Not Loading

1. Check file location: `~/.config/c9s/config.toml`
2. Validate TOML syntax (run through a TOML parser)
3. Check file permissions (must be readable)
4. c9s silently ignores invalid configs and uses defaults

### Hotkey Not Working

1. Ensure the binding name is correct (see default keymap)
2. Check key format matches supported syntax
3. Some keys may be intercepted by terminal or OS
4. Use lowercase for Ctrl combinations (`"ctrl+c"`, not `"Ctrl+C"`)

### Alias Not Resolving

1. Aliases only work in palette mode (`:command`)
2. Check alias name doesn't conflict with built-in commands
3. Verify aliases.toml syntax with TOML parser
