# Containers screen (`:containers`, `:c`)

The home screen of c9s. Lists containers managed by Apples container runtime, refreshes every 2 seconds, and supports keyboard-driven lifecycle operations.

## Columns

- **SHORT-ID** — first 12 hex chars of the container UUID.
- **IMAGE** — image reference (e.g. nginx:1.27).
- **STATE** — running stopped paused stopping created. Color-coded.
- **UPTIME** — duration since the container started.
- **CPU** — configured CPU count.
- **MEM** — configured memory limit.

## Hotkeys

| Key | Action |
|---|---|
| up down or j k | Move selection |
| g G | Top bottom |
| space | Toggle mark on focused row |
| asterisk | Mark all visible |
| Esc | Clear marks dismiss filter close modal |
| slash | Filter by image or short-id |
| r | Refresh |
| question | Help overlay |
| colon | Command palette |
| s | Suspend c9s and shell into the container |
| d | Inspect JSON viewer |
| x | Stop |
| ShiftK | Remove container |
| ShiftR | Restart |
| ShiftD | Delete with confirmation |
| p | Pause unpause greyed out |
