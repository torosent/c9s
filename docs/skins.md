# Skins

> **Note:** Full skin system implementation is planned for a future release. This document describes the intended design.

## Overview

c9s will support customizable color themes ("skins") to personalize the appearance of the TUI.

## Bundled Skins

Planned bundled skins:
- **dark** (default): GitHub-style dark theme
- **light**: Light theme for bright environments
- **k9s-dark**: k9s-compatible dark theme
- **k9s-light**: k9s-compatible light theme

## Custom Skins

### Creating a Custom Skin

Create a TOML file in `~/.config/c9s/skins/<name>.toml`:

```toml
# ~/.config/c9s/skins/my-theme.toml
[colors]
fg = "#e0e0e0"
bg = "#1a1a1a"
border = "#404040"
accent = "#00d4ff"
dim = "#808080"
success = "#00ff00"
warning = "#ffaa00"
error = "#ff0000"
selection_fg = "#ffffff"
selection_bg = "#0066cc"
header_fg = "#f0f0f0"
header_bg = "#2a2a2a"

[state_colors]
running = "#00ff00"
exited = "#808080"
paused = "#ffaa00"
stopping = "#ff0000"
created = "#00d4ff"
```

### Applying a Skin

**Via config.toml:**
```toml
[theme]
name = "my-theme"
```

**Via palette command (live):**
```
:skin my-theme
```

## k9s Skin Importer

> **Note:** Planned feature for future release.

Convert k9s skin files to c9s format:

```bash
c9s import-skin ~/.config/k9s/skins/dracula.yaml
```

Or via palette command:
```
:import-skin ~/.config/k9s/skins/dracula.yaml
```

The importer will map k9s color keys to c9s palette keys and save to `~/.config/c9s/skins/`.

### k9s Compatibility

k9s skins use YAML format. The importer maps these keys:

| k9s Key | c9s Key |
|---------|---------|
| `body.fgColor` | `fg` |
| `body.bgColor` | `bg` |
| `body.logoColor` | `accent` |
| `info.fgColor` | `info_fg` |
| `info.sectionColor` | `info_section` |
| `frame.title.fgColor` | `header_fg` |
| `frame.title.bgColor` | `header_bg` |
| `frame.border.fgColor` | `border` |

## Color Format

Skins support standard CSS color formats:
- Hex: `"#ff0000"`, `"#f00"`
- Named colors: `"red"`, `"blue"`, `"green"`
- RGB: `"rgb(255, 0, 0)"`

## Skin Location

Skins are searched in this order:
1. `~/.config/c9s/skins/<name>.toml` (user skins)
2. Built-in bundled skins (embedded in binary)

## Live Preview

> **Note:** Planned feature.

Use `:skin <name>` to preview skins without restarting. The change applies immediately to all screens.

To make permanent, update `config.toml`:
```toml
[theme]
name = "preferred-skin"
```

## Creating Accessible Skins

When designing custom skins, consider:
- **Contrast:** Ensure sufficient contrast between fg/bg (WCAG AA: 4.5:1 for text)
- **Color Blindness:** Don't rely solely on color to convey state
- **Low Vision:** Use larger contrast ratios for better readability

## Examples

### Minimal Custom Skin

```toml
# ~/.config/c9s/skins/minimal.toml
[colors]
fg = "#ffffff"
bg = "#000000"
accent = "#00ffff"
```

Unspecified colors fall back to defaults.

### High Contrast Skin

```toml
# ~/.config/c9s/skins/high-contrast.toml
[colors]
fg = "#ffffff"
bg = "#000000"
border = "#ffffff"
accent = "#ffff00"
dim = "#aaaaaa"
success = "#00ff00"
warning = "#ffaa00"
error = "#ff0000"
selection_fg = "#000000"
selection_bg = "#ffff00"
```

## Troubleshooting

### Skin Not Found

1. Check file exists: `ls ~/.config/c9s/skins/<name>.toml`
2. Verify TOML syntax
3. Check theme name in config matches file basename (without `.toml`)
4. Bundled skins don't require files (dark, light, k9s-dark, k9s-light)

### Colors Not Applying

1. Ensure terminal supports 24-bit color (truecolor)
2. Check `$COLORTERM` environment variable (should be `truecolor` or `24bit`)
3. Some terminals require explicit configuration for truecolor support
4. Test with: `echo $COLORTERM`

### k9s Import Fails

1. Verify k9s skin file is valid YAML
2. Check file permissions
3. Ensure destination directory exists: `~/.config/c9s/skins/`
4. k9s skin must have `.yaml` or `.yml` extension

## Future Enhancements

Planned features:
- Per-screen skin overrides
- Dynamic skin switching based on time of day
- Skin preview gallery in TUI
- More bundled skins (Dracula, Nord, Solarized, etc.)
- Gradient support for backgrounds
- Font weight and style customization
