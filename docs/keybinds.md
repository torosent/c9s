# Keybinds

This page lists the default keybinds. Plan 7 will auto-generate it from `Hotkeys()` declarations in the codebase.

## Global

| Key | Action |
|---|---|
| `:` | Command palette |
| `?` | Help overlay (lists current screen's keys) |
| `Ctrl+E` | Toggle header |
| `Ctrl+C` | Quit |
| `r` | Refresh active screen |
| `/` | Filter |
| `space` | Mark focused row |
| `*` | Mark all visible rows |
| `Esc` | Clear marks / dismiss filter / close modal |

## Per-screen

| Screen | Page |
|---|---|
| Containers (`:c`) | [containers.md](screens/containers.md) |
| Images (`:i`) | [screens/images.md](screens/images.md) |
| Volumes (`:v`) | [screens/volumes.md](screens/volumes.md) |
| Networks (`:n`) | [screens/networks.md](screens/networks.md) |
| Builder (`:b`) | [screens/builder.md](screens/builder.md) |
| Registry (`:reg`) | [screens/registry.md](screens/registry.md) |
| System (`:sys`) and sub-screens | [screens/system.md](screens/system.md) |

## Modals

| Modal | Page |
|---|---|
| Run form (`:run`, or `R` on Images) | [screens/run.md](screens/run.md) |
| Build form (`:build`) | [screens/build.md](screens/build.md) |
| Login (`:login`, or `L` on Registry) | inline in [screens/registry.md](screens/registry.md) |
| Generic text input (`:create`, `:tag`, `:save`, `:load`) | label-routed `TextInputResultMsg` |

## Palette commands

| Command | Effect |
|---|---|
| `:c` / `:containers` | Switch to Containers |
| `:i` / `:images` | Switch to Images |
| `:v` / `:volumes` | Switch to Volumes |
| `:n` / `:networks` | Switch to Networks |
| `:b` / `:builder` | Switch to Builder |
| `:reg` / `:registry` | Switch to Registry |
| `:sys` / `:system` | Switch to System services |
| `:df` | System disk usage |
| `:dns` | System DNS domains |
| `:property` | System properties |
| `:kernel` | Kernel configuration |
| `:logs` | System log stream |
| `:run [image]` | Open the run form |
| `:build [path]` | Open the build form |
| `:login [host]` | Open the registry login form |
| `:create [name]` | Generic text-input prompt |
| `:tag [src]` | Tag prompt |
| `:save [ref] [tar]` | Save image prompt |
| `:load [tar]` | Load image prompt |
| `:pull <ref>` | Open pull progress modal |
| `:push <ref>` | Open push progress modal |
| `:help`, `:?` | Help overlay |
| `:q`, `:quit`, `:exit` | Quit |

