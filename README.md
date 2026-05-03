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
- 🔌 **Plugin system** — define your own hotkeys + shell commands per resource
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

## Development

```bash
make install-tools   # gofumpt, staticcheck, golangci-lint
make ci              # format check + lint + race tests + coverage gate
```

## Documentation

- [Design spec](docs/superpowers/specs/2026-05-02-c9s-design.md)
- [Architecture overview](docs/architecture.md)
- [Contributing guide](docs/contributing.md)

## Versioning

c9s follows [Semantic Versioning](https://semver.org/). Pre-1.0 releases (`0.x.y`) may include breaking changes between minor versions; we'll always note these in the [changelog](CHANGELOG.md).

- `0.x.y` — pre-1.0 development; minor versions may break.
- `1.0.0+` — stable; breaking changes only on major bumps.

## Contributing

See [docs/contributing.md](docs/contributing.md) for development setup, testing, and how to release.

## License

[Apache 2.0](LICENSE)
