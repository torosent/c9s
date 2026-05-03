# Images screen (`:images`, `:i`)

Lists images known to the local Apple `container` runtime, refreshing every 2 seconds.

## Columns

- **REPOSITORY** — the image repository (e.g. `ghcr.io/acme/api`); `<none>` for dangling images.
- **TAG** — the tag part of the reference (e.g. `1.4.2`); `<none>` for dangling.
- **ID** — first 12 hex chars after the `sha256:` prefix.
- **CREATED** — short relative time (e.g. `3d`, `2h`, `45s`).
- **SIZE** — human-readable byte total.

## Hotkeys

| Key | Action |
|---|---|
| `↑/↓` or `j/k` | Move selection |
| `g/G` | Top / bottom |
| `space` | Toggle mark on focused row |
| `*` | Mark all visible |
| `Esc` | Clear marks |
| `/` | Filter by repository, tag, reference, or short-id |
| `r` | Refresh |
| `?` | Help overlay |
| `:` | Command palette |
| `d` | Inspect JSON viewer for the focused image |
| `t` | Tag — opens a TextInput prompt for the new tag |
| `P` | Push — opens the progress modal streaming `container image push` |
| `D` | Delete with confirmation (multi-select via marks) |
| `R` | Run-from-image — opens the run form pre-filled with this reference |

## Palette commands

- `:pull <ref>` — opens the pull progress modal.
- `:build <path>` — opens the build form (defaults path to the current directory).
- `:load <tar>` — load an image tarball back into the runtime.
- `:save <ref> <tar>` — save the focused image to a tarball.
- `:prune` (Plan 5) — remove unused images.
