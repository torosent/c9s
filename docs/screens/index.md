# Screens overview

c9s organizes its UI around resource screens. Switch between them via the
command palette (`:`) — each screen has a long name (e.g. `:containers`)
and a single-letter alias (`:c`).

## Resource screens

| Palette | Alias | Description |
|---------|-------|-------------|
| [`:containers`](containers.md) | `:c` | Lifecycle management for running containers (start, stop, exec, logs, inspect). |
| [`:images`](images.md) | `:i` | Local image registry. Tag, push, delete, run-from-image. |
| [`:volumes`](volumes.md) | `:v` | Persistent volumes. Create, delete, prune. |
| [`:networks`](networks.md) | `:n` | Container networks. Create, delete, prune. |
| [`:builder`](builder.md) | `:b` | Builder daemon status. |
| [`:registry`](registry.md) | `:reg` | Registry credentials. Login, logout, set-default. |
| [`:system`](system.md) | `:sys` | System services + sub-screens (`:df`, `:dns`, `:property`, `:kernel`, `:logs`). |

## k9s-parity screens

These mirror k9s features for users migrating from Kubernetes.

| Palette | Description |
|---------|-------------|
| `:jobs` | Background streaming jobs (builds, pulls, pushes detached with Ctrl+Z). |
| `:pinned` | Cross-screen bookmarks. Press `b` on any row to pin it. |
| `:errors` | Persistent error log viewer. |
| `:pulses` | Live health dashboard (services, builder, resource counts, sparkline, errors). |
| `:xray` | Tree view of resource relationships (image → containers → networks/volumes). |

## Action screens

These are not table-based; they handle specific workflows.

| Palette | Description |
|---------|-------------|
| [`:run <image>`](run.md) | Multi-field form to launch a container. |
| [`:build <path>`](build.md) | Multi-field form to build an image. |
| `:logs <id>` | Open the log viewer for a single container. |
| `:pull <ref>` / `:push <ref>` | Direct streaming progress for image transfers. |

## Common hotkeys

Most table screens share a baseline of bindings:

- `↑` / `↓` or `j` / `k` — move selection
- `g` / `G` — top / bottom
- `space` — toggle mark on focused row
- `*` — mark all visible
- `Esc` — clear marks / dismiss filter / close modal
- `/` — filter rows by free-text
- `r` — refresh
- `?` — help overlay
- `:` — command palette
- `Shift+S` — open the sort picker
- `b` — pin focused row to `:pinned`

Per-screen hotkeys are documented on each screen's page.
