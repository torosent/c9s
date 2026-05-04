# c9s

A k9s-style terminal UI for managing [Apple's `container` runtime](https://github.com/apple/container) on macOS.

![c9s demo](docs/assets/c9s.gif)

## Features

- 🚀 **Fast, keyboard-driven interface** inspired by k9s
- 📦 **Manage containers, images, volumes, networks** + builder, registry, system + sub-screens
- 🎨 **Customizable themes** (TOML skins, including k9s skin importer)
- 📜 **Multi-source log streaming** with ring buffer, follow-tail, save, and level coloring
- 🔨 **Background jobs** for builds, pulls, pushes — `Ctrl+Z` detach, `:jobs` re-attach
- 🖱️ **Mouse support** (click row, click header to sort, scroll wheel)
- 📊 **k9s parity** — `:pulses` dashboard, `:xray` resource graph, `:pinned`, `:errors`
- 🛡️ **Read-only mode** — `--readonly` hides destructive bindings

## Requirements

- Apple Silicon Mac
- [Apple `container`](https://github.com/apple/container/releases) 0.12 or later

## Install

### Homebrew

```bash
brew tap torosent/c9s
brew install c9s
```

### Direct download

Download the Apple Silicon archive from the [latest release](https://github.com/torosent/c9s/releases/latest), extract, and put `c9s` on your PATH:

```bash
curl -L -o c9s.tar.gz https://github.com/torosent/c9s/releases/latest/download/c9s_0.1.0_darwin_arm64.tar.gz
tar -xzf c9s.tar.gz
sudo mv c9s /usr/local/bin/
c9s --version
```

### Build from source

```bash
git clone https://github.com/torosent/c9s.git
cd c9s
make build
./bin/c9s --version
```

For development setup (linters, formatters, tests), see [docs/contributing.md](docs/contributing.md).

## Quick start

After building, launch:

```bash
./bin/c9s
```

You'll land on the **containers** screen. Press `?` to see the keybinds. Type `:q` and Enter (or press `Ctrl+C`) to quit.

## Commands

c9s is keyboard-driven. Press `:` to open the command palette, then type any of the commands below (Tab autocompletes). All commands live in `internal/ui/palette/catalog.go` if you want the canonical list.

### Switch screens

| Command | Alias | What it does |
|---|---|---|
| `:containers` | `:c` | Container list (default screen) |
| `:images` | `:i` | Image list |
| `:volumes` | `:v` | Volume list |
| `:networks` | `:n` | Network list |
| `:builder` | `:b` | Builder status |
| `:registry` | `:reg` | Configured registry logins |
| `:system` | `:sys` | System services overview |

### System sub-screens

| Command | What it does |
|---|---|
| `:df` | Disk usage breakdown |
| `:dns` | DNS domains (CRUD) |
| `:property` | System properties (edit / reset) |
| `:kernel` | Kernel-only properties (read-only viewport) |
| `:logs` | Live `container system logs --follow` |

### k9s parity

| Command | What it does |
|---|---|
| `:pulses` | Live health dashboard |
| `:xray` | Resource relationships tree |
| `:jobs` | Background streaming jobs (build / pull / push) |
| `:pinned` | Bookmarked resources |
| `:errors` | Error log viewer |

### Actions

| Command | Args | What it does |
|---|---|---|
| `:run` | `<image>` | Launch a new container (run modal) |
| `:build` | `<path>` | Build an image from a directory |
| `:pull` | `<ref>` | Pull image from registry |
| `:push` | `<ref>` | Push image to registry |
| `:create` | `<name>` | Create resource on the active screen |
| `:tag` | `<src> <dst>` | Tag an image |
| `:save` | `<ref> <tar>` | Save image to a tar file |
| `:load` | `<tar>` | Load image from a tar file |
| `:login` | `<host>` | Open the registry login modal |
| `:acr-login` | `<registry>` | Azure Container Registry login via Azure AD ([details](docs/screens/registry.md#azure-container-registry-acr)) |

### Config & customization

| Command | Args | What it does |
|---|---|---|
| `:skin` | `<name>` | Switch to a TOML skin (try `:skins` to list) |
| `:skins` | — | Open an interactive skin picker |
| `:import-skin` | `<path>` | Import a k9s YAML skin |
| `:install-docker-shim` | `[path]` | Install a `docker → container` compatibility shim ([details](docs/docker-shim.md)) |
| `:uninstall-docker-shim` | `[path]` | Remove the c9s docker shim |

### Meta

| Command | Alias | What it does |
|---|---|---|
| `:help` | `:?` | Show help overlay for the active screen |
| `:quit` | `:q`, `:exit` | Quit c9s |

`c9s` also exposes two sibling subcommands of its own (i.e. you run them in your shell, not inside the TUI):

```bash
c9s --version
c9s import-skin <path/to/k9s-skin.yaml>      # import a k9s skin
c9s install-docker-shim [--path P] [--force] # install docker→container shim
c9s uninstall-docker-shim [--path P]         # remove the shim
```

## Development

```bash
make install-tools   # gofumpt, staticcheck, golangci-lint
make ci              # format check + lint + race tests + coverage gate
```

## Documentation

- [Documentation site](https://torosent.github.io/c9s)
- [Architecture overview](docs/architecture.md)
- [Quick start](docs/quick-start.md)
- [Configuration](docs/configuration.md)
- [Skins](docs/skins.md)
- [Hotkeys reference](docs/hotkeys.md)
- [k9s migration](docs/k9s-migration.md)
- [Contributing guide](docs/contributing.md)

## Versioning

c9s follows [Semantic Versioning](https://semver.org/). Pre-1.0 releases (`0.x.y`) may include breaking changes between minor versions; we'll always note these in the [changelog](CHANGELOG.md).

- `0.x.y` — pre-1.0 development; minor versions may break.
- `1.0.0+` — stable; breaking changes only on major bumps.

## Contributing

See [docs/contributing.md](docs/contributing.md) for development setup, testing, and how to release.

## License

[Apache 2.0](LICENSE)
