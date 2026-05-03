# Networks screen (`:networks`, `:n`)

Lists networks managed by the local Apple `container` runtime, refreshing every 2 seconds.

## Columns

- **NAME** — network name.
- **DRIVER** — backing driver (typically `bridge`).
- **SUBNET** — assigned CIDR (`-` when unset).
- **CONTAINERS** — comma-separated list of containers attached.

## Hotkeys

| Key | Action |
|---|---|
| `↑/↓` or `j/k` | Move selection |
| `space` / `*` | Mark / mark all |
| `Esc` | Clear marks |
| `/` | Filter by name, driver, or subnet |
| `r` | Refresh |
| `d` | Inspect |
| `D` | Delete with confirmation |

## Palette commands

- `:create <name>` — create a new network.
- `:prune` — remove unused networks (Plan 5).
