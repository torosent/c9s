# Builder screen (`:builder`, `:b`)

Single-card view of the Apple `container` build subsystem. Refreshes every 2 seconds.

## Card fields

- **STATE** — `running`, `stopped`, `error`, or `unknown`. Color-coded.
- **CPU** — CPU count assigned to the builder VM.
- **MEM** — memory size (human-readable).
- **UPTIME** — how long the builder has been running.

## Hotkeys

| Key | Action |
|---|---|
| `S` | Start the builder |
| `X` | Stop the builder |
| `D` | Delete the builder VM (confirm) |
| `r` | Refresh status |
| `?` | Help overlay |

## Notes

- Deleting the builder destroys the build cache. The confirmation modal makes
  the destructive nature explicit.
- Errors during `start`/`stop`/`delete` surface as toasts in the status bar.
