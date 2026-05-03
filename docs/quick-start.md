# Quick Start

## Installation

### From Source

```bash
git clone https://github.com/torosent/container-tui.git
cd container-tui
make build
sudo mv bin/c9s /usr/local/bin/
```

### From Release

Download the latest release from [GitHub Releases](https://github.com/torosent/container-tui/releases).

## Running c9s

### With Real Container Runtime

```bash
c9s
```

c9s will connect to your Apple Containers CLI automatically.

### Demo Mode

Try c9s without a container runtime using the `--demo-data` flag:

```bash
c9s --demo-data
```

This spins up a populated fake client with sample containers, images, volumes, and networks.

## Navigation Basics

- **Arrow keys** or `j`/`k` to navigate lists
- **Enter** to select or open details
- **`q`** to quit
- **`:`** to open the command palette
- **`?`** to see help/hotkeys
- **`/`** to filter items
- **Space** to mark items for batch operations
- **Mouse** clicking and scrolling supported

## Command Palette

Press `:` to open the command palette and type:

- `:containers` - Switch to containers screen
- `:images` - Switch to images screen  
- `:volumes` - Switch to volumes screen
- `:networks` - Switch to networks screen
- `:builder` - Switch to builder screen
- `:registry` - Switch to registry screen
- `:system` - Switch to system services screen
- `:jobs` - View background jobs
- `:errors` - View error logs
- `:pinned` - View pinned items

## Next Steps

- Browse the [Screens](screens/index.md) documentation to learn about each screen
- Check out [Hotkeys](hotkeys.md) for the full keyboard reference
- Read [Configuration](configuration.md) to customize c9s
- Explore [Skins](skins.md) to change the UI theme
