# FAQ

## General

### Why c9s?

c9s (pronounced "sinks") brings the k9s experience to Apple Containers. It provides a fast, keyboard-driven TUI for managing containers, images, volumes, networks, and more—all without leaving your terminal.

### Does c9s work without the Apple Containers CLI?

Yes! Use `c9s --demo-data` to explore the interface with populated sample data, even if you don't have the container runtime installed.

### Can I use my k9s skin?

Yes! c9s supports k9s skin format. Import your k9s skin with:

```bash
c9s import-skin ~/.config/k9s/skins/my-skin.yaml
```

See [Skins](skins.md) for more details.

### How do I add a plugin?

Plugins are defined in `~/.config/c9s/plugins.yaml`. See [Configuration](configuration.md#plugins) for the format.

### Does c9s support remote container hosts?

c9s uses the Apple Containers CLI under the hood, so if your CLI is configured to connect to a remote host, c9s will use that connection automatically.

## Usage

### How do I delete multiple containers at once?

1. Press **Space** to mark containers
2. Press **Shift+D** to delete all marked containers
3. Confirm the deletion

This batch operation pattern works across all screens (containers, images, volumes, etc.).

### How do I view logs from multiple containers?

1. Mark containers with **Space**
2. Press **`l`** to open the multi-source log viewer
3. Logs from all marked containers stream in parallel with color-coded prefixes

### Can I detach from a running build or pull?

Yes! When a build or pull is running:

1. Press **Ctrl+Z** to detach
2. The job continues in the background
3. Switch to `:jobs` to see all background jobs
4. Press **Enter** on a job to re-attach

### How do I sort a list?

On any sortable screen (containers, images, volumes, networks, jobs):

1. Press **`s`** to open the sort modal
2. Select the column to sort by
3. Choose ascending or descending order

Alternatively, click on column headers with your mouse.

## Troubleshooting

### c9s shows "container CLI not found"

Ensure the Apple Containers CLI is installed and in your PATH. You can test with:

```bash
which container
container version
```

### c9s crashes on startup

Check the error logs:

```bash
cat ~/.local/state/c9s/errors/*.log
```

Or launch c9s and press `:errors` to view the errors screen.

### Logs aren't streaming

Try increasing the refresh interval in `~/.config/c9s/config.yaml`:

```yaml
ui:
  refresh_interval: 2s
```

### Mouse isn't working

Ensure your terminal supports mouse events. Most modern terminals (iTerm2, Alacritty, Wezterm, etc.) support this by default.

## Development

### How do I build from source?

```bash
git clone https://github.com/torosent/container-tui.git
cd container-tui
make build
```

The binary will be in `bin/c9s`.

### How do I run tests?

```bash
make test
```

Or with race detection and coverage:

```bash
make ci
```

### How do I contribute?

See [Contributing](contributing.md) for guidelines on filing issues, submitting PRs, and setting up your development environment.
