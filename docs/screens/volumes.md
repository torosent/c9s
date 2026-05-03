# Volumes screen (`:volumes`, `:v`)

Lists volumes managed by the local Apple `container` runtime, refreshing every 2 seconds.

## Columns

- **NAME** — volume name.
- **DRIVER** — backing driver (typically `local`).
- **MOUNTPOINT** — host directory backing the volume.
- **SIZE** — human-readable byte total (`-` if unknown).
- **USED-BY** — comma-separated list of containers attached to the volume.

## Hotkeys

| Key | Action |
|---|---|
| `↑/↓` or `j/k` | Move selection |
| `space` / `*` | Mark / mark all |
| `Esc` | Clear marks |
| `/` | Filter by name, driver, or mountpoint |
| `r` | Refresh |
| `d` | Inspect (JSON dump of the focused volume) |
| `D` | Delete with confirmation (multi-select via marks) |

## Palette commands

- `:create <name>` — create a new volume.
- `:prune` — remove unused volumes (Plan 5).
