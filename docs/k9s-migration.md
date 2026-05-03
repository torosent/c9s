# Migrating from k9s

c9s is intentionally modeled on [k9s](https://k9scli.io/) — most of the
keyboard navigation, screen organization, and configuration surface should
feel familiar. This page maps k9s concepts to c9s.

## Conceptual mapping

| k9s concept | c9s equivalent |
|-------------|----------------|
| Cluster | Apple `container` runtime (single host) |
| Namespace | (none — flat) |
| Pod | Container (`:containers`) |
| Image (used by pod) | Image (`:images`) |
| PVC / PV | Volume (`:volumes`) |
| Service / Ingress | Network (`:networks`) |
| Node | Builder + system services (`:builder`, `:system`) |
| Context | (none — there is one host) |
| Skin (YAML) | Skin (TOML) — see [Skins](skins.md) |
| `pulses` | `:pulses` |
| `xray` | `:xray` |
| `popeye` | (not implemented) |

## Hotkey parity

| Action | k9s | c9s |
|--------|-----|-----|
| Command palette | `:` | `:` |
| Help overlay | `?` | `?` |
| Quit | `Ctrl+C` / `:q` | `Ctrl+C` / `:q` |
| Filter | `/` | `/` |
| Mark / unmark | `space` | `space` |
| Mark all | `Ctrl+a` | `*` |
| Sort | `Shift+S` | `Shift+S` |
| Toggle header | `Ctrl+E` | `Ctrl+E` |
| Pin | `Ctrl+P` | `b` |

## Skin migration

c9s ships with `k9s-dark.toml` and `k9s-light.toml` skins that mirror
k9s's bundled themes. To bring a custom k9s skin over:

```bash
c9s import-skin /path/to/your-k9s-skin.yml
```

This writes `~/.config/c9s/skins/your-k9s-skin.toml`. Activate with:

```
:skin your-k9s-skin
```

## What's different

- **No namespaces.** c9s targets a single Apple `container` host. There is
  no notion of context-switching or namespace filtering.
- **Suspend-and-shell-out for exec.** Like k9s's `s` shortcut, c9s suspends
  itself, runs `container exec -it <id> $SHELL`, and resumes when you exit.
- **Streaming jobs.** Build / pull / push run in a progress modal that you
  can detach (`Ctrl+Z`) into the `:jobs` panel; from there, `Enter`
  re-attaches.
- **Capability gating.** Pause/unpause is hidden (or annotated) when the
  underlying `container` binary doesn't support it. Future capabilities
  follow the same pattern.
- **Read-only mode.** Run with `--readonly` (or `[ui] readonly = true` in
  config.toml) to hide destructive bindings.

## Configuration files

| k9s file | c9s file |
|----------|----------|
| `~/.config/k9s/config.yaml` | `~/.config/c9s/config.toml` |
| `~/.config/k9s/hotkeys.yaml` | `~/.config/c9s/hotkeys.toml` |
| `~/.config/k9s/aliases.yaml` | `~/.config/c9s/aliases.toml` |
| `~/.config/k9s/views.yaml` | `~/.config/c9s/views.toml` |
| `~/.config/k9s/skins/*.yaml` | `~/.config/c9s/skins/*.toml` |

YAML in k9s, TOML in c9s. The schema is similar but not identical — see
[Configuration](configuration.md) for the full reference.

## What's not (yet) implemented

These k9s features aren't in c9s today:

- `popeye` (cluster sanitizer) — not applicable.
- Live RBAC editor — not applicable.
- Multi-cluster management — c9s is single-host.

Open an [issue](https://github.com/torosent/c9s/issues) if there's a
specific k9s feature you want surfaced in c9s.
