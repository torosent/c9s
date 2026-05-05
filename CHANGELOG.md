# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- **Post-exec rendering corruption** (the "exit shell shows only 3
  banner rows" bug). After `tea.ExecProcess` restored the terminal on
  macOS, stdout was sometimes left in non-blocking mode. The bubbletea
  v2 renderer issues full-frame writes (~10 KB each) and the kernel
  TTY buffer caps at ~1 KB, so the very first post-resume write
  returned `EAGAIN` after only 1024 bytes. The renderer treats partial
  writes as fatal, drops the rest of the frame, and never recovers
  (subsequent renders are skipped because the cached `lastView`
  matches the new view). The fix wraps `os.Stdout` in a small
  `blockingwriter` that retries on `EAGAIN`/`EWOULDBLOCK` so the full
  frame always reaches the terminal.

### Changed

- **Upgraded to bubbletea v2 (`charm.land/bubbletea/v2`)**, plus
  matching v2 of `lipgloss` and `bubbles`. v2 ships a new cell-based
  renderer (uses Charm's `ultraviolet` terminal library) that
  correctly handles `tea.ExecProcess` resume — the old line-diff
  renderer in v1.3.10 had a bug where the `lastRenderedLines` cache
  survived the suspend/resume cycle even though `repaint()` was
  called, leaving the user with banner-bottom + one container row +
  acres of blank space after `s` → bash → `exit`. v2's renderer
  doesn't have this bug. Instrumented byte-stream capture confirms
  the post-exec frame is now drawn correctly.
  - All key handlers updated to `tea.KeyPressMsg` (v2's
    `tea.KeyMsg` is now an interface).
  - All mouse handlers updated to `tea.MouseClickMsg` /
    `tea.MouseWheelMsg` (also interfaces in v2).
  - Root `Model.View()` returns `tea.View` instead of `string`;
    altscreen and mouse mode are now declared via `View.AltScreen`
    and `View.MouseMode` rather than `tea.NewProgram` options.
  - `lipgloss.Color` is a constructor function in v2; the `Palette`
    struct fields are now `image/color.Color`.

### Added

- **Shell picker modal** — `s` on a running container now opens a
  small modal asking whether to use `/bin/bash` or `/bin/sh` rather
  than blindly using the host's `$SHELL`. The host shell (often
  `/bin/zsh` on macOS) is rarely present inside Linux containers,
  and Apple's `container` returns exit 0 even when exec fails, so a
  missing shell would silently leave the user staring at a glitched
  half-rendered TUI. Press `b`/`s` for a one-keystroke pick or use
  arrow keys + Enter.

### Fixed

- **`ShellPickedMsg` was swallowed by the still-open picker modal.**
  The picker batches `ShellPickedMsg` alongside `CloseModalMsg`, but
  `tea.Batch` doesn't guarantee ordering. When `ShellPickedMsg`
  arrived first the picker was still top of stack, the modal received
  the message, didn't handle it, and the user's pick vanished — the
  classic "I clicked bash and nothing happened" symptom. Added an
  explicit typed case in `app.Update` that forwards `ShellPickedMsg`
  directly to the active screen, mirroring `ConfirmResultMsg`.
- **Probe shell existence before suspending the TUI.** Apple's
  `container exec -it <id> <shell>` returns **exit 0 even when the
  shell isn't installed** — it writes the error to stderr (visible
  for milliseconds before altscreen re-entry hides it) and exits.
  `tea.ExecProcess` sees a clean exit so we can't surface a useful
  toast post-hoc. Now probe `container exec <id> test -x <shell>`
  (3-second timeout, no `-it`) before running the interactive exec;
  if the probe fails we toast `<shell> not available in <id> — try
  the other shell` and skip the suspend entirely.
- **Glitched TUI after `tea.ExecProcess` returns.** Even on a
  successful shell session, after exit the next altscreen frame
  sometimes rendered on top of stale cells (truncated table +
  leftover output visible). `tea.WindowSize()` alone wasn't enough
  because bubbletea's renderer preserves cells it thinks are
  unchanged. The handler now batches `tea.ClearScreen` (emits
  `\033[2J\033[H`) ahead of `tea.WindowSize()` to force a full
  altscreen repaint.
- **`x` (stop), `Shift+K` (kill), `Shift+R` (restart), and `p` (pause)
  now refresh the table immediately.** Previously they relied on the
  2-second poll tick, so the user pressed `x` to stop a container and
  saw `running` for up to 2 seconds. Each lifecycle action now batches
  a follow-up `ListContainers` refresh, mirroring the existing
  `delete` / `prune` behaviour, and surfaces a clear "stopped <id>" /
  "killed <id>" / etc. toast.
- **`s` (shell) on a non-running container shows a toast instead of
  failing silently.** `container exec -it <id> <shell>` exits
  immediately when the target container isn't running, leaving the
  user staring at the same screen with no feedback. The screen now
  refuses to issue the exec for non-running containers and surfaces
  `can't open shell: <id> is stopped`. ExecProcess errors at the
  `app.go` layer are also surfaced as a toast so any other failure
  (image lacks `/bin/sh`, race with another stop, etc.) is visible.
- **Splash dropping the active screen's first refresh and tick.** The
  app's catch-all message-forwarding block was gated behind
  `!m.showSplash`, which meant any message dispatched by the active
  screen's `Init` (the immediate `RefreshedMsg` plus the first
  `TickMsg`) was silently dropped while the splash was still on
  screen. Combined with `clock.Real().Tick()` being one-shot via
  `time.After`, this killed the auto-refresh loop entirely — users
  saw `0 items` until they happened to switch screens or press `r`.
  Now `RefreshedMsg`, `TickMsg`, and other internal messages are
  forwarded to the active screen even while the splash is up; only
  KeyMsg / WindowSizeMsg are still routed to the splash. Regression
  test in `internal/ui/app_test.go`.

## [0.1.4] - 2026-05-04

### Added

- **`Shift+P` / `:prune`** on the containers screen — removes all
  stopped containers via Apple's `container prune`. Opens a confirm
  modal listing the stopped containers (short-id, image, state tag) so
  users can sanity-check before deleting; surfaces a clear status
  toast and refreshes the list afterwards. If there are no stopped
  containers, shows a "no stopped containers to prune" toast instead
  of an empty modal.
- **Help overlay now lists `sort` (`Shift+S`) and `screen_switch`
  (`1-9, 0`).** Both bindings were already wired in `app.go` but not
  registered in `keymap.Default()`, so `?` never showed them. Added
  as catalog entries; functional dispatch is unchanged.

### Fixed

- **Containers table header / data column alignment.** Bubbles' default
  `Header` style applies `Padding(0, 1)`; we had explicitly cleared
  `Cell` padding (intentional, so per-cell backgrounds don't fight the
  selected-row highlight) but left it inherited on `Header`. Each
  header cell rendered 2 cols wider than its data cell — across the
  7-column containers table that's +14 cols, enough to wrap the
  header onto two lines and shift every label out of alignment with
  the data column beneath at typical terminal widths. Now
  `Padding(0, 0)` on both, with inter-column gap provided by the
  leading-space prefix on every column title and value.
- **Help-overlay column alignment.** The 2-column layout used
  `%-40s  %s` to pad the *combined* "label + keys" string, so right-
  side cells ended at variable columns depending on key-string length
  (`[s]` vs `[shift+p, P]` is 9 chars apart). Rewrote to compute max
  label/keys widths from the data and render
  `padRight(label, maxL) + " " + padRight(keys, maxK)` with a fixed
  separator; every right-side `[` now sits at a stable column.

## [0.1.3] - 2026-05-04

### Added

- **`:acr-login <registry>` palette command.** One-step Azure Container
  Registry login that wraps the Microsoft-recommended AAD-token recipe:
  shells out to `az acr login --expose-token`, then pipes the token
  into `container registry login` as the zero-GUID anonymous user via
  `--password-stdin`. Accepts either bare names (`myreg`) or full
  hostnames (`myreg.azurecr.io`); sovereign clouds (`*.azurecr.us`,
  `*.azurecr.cn`) preserved. Failure cases include `az` not on PATH,
  empty tokens, and `az acr login` errors — all surfaced via a new
  dismissable info modal with concrete remediation hints.
- **Docker compatibility shim** — a POSIX bash script that maps the
  most common `docker(1)` verbs to their Apple `container` CLI
  equivalents. Install via `c9s install-docker-shim` (defaults to
  `~/.local/bin/docker`) or the `:install-docker-shim` palette
  command; uninstall is sentinel-protected so it can't accidentally
  delete a real docker binary. The shim translates
  `ps`/`images`/`rm`/`rmi`/`pull`/`push`/`tag`/`volume ls`/`network ls`/
  `login`/`logout`/`info` and falls through unhandled verbs. See
  `docs/docker-shim.md`.
- **Existing-docker detection.** Before installing the shim, c9s walks
  PATH and reports any pre-existing `docker` executables (in PATH
  order) plus whether Docker Desktop is installed at
  `/Applications/Docker.app`, so users know exactly what they're
  competing with.
- **Dismissable info modal** (`modals.NewInfo`) with OK/Warning/Error
  levels — used by `:acr-login` and `:install/uninstall-docker-shim`
  for clear, non-truncated result feedback.
- **`acr-login`/`install-docker-shim`/`uninstall-docker-shim`** entries
  in the palette catalog so Tab autocomplete surfaces them.

### Fixed

- `cli.DefaultCtx` no longer trips `go vet`'s lostcancel analyzer. A
  one-line watcher goroutine calls the cancel returned by WithTimeout
  when the deadline fires, satisfying the analyzer while keeping call
  sites at the one-arg form they had after the v0.1.1 timeout audit.

## [0.1.2] - 2026-05-03

### Changed

- Default theme is now `k9s-gruvbox-dark`. Previously new installations
  rendered with the built-in `dark` palette; the gruvbox-dark skin
  ships bundled and is more pleasant out of the box. Existing users
  whose `~/.config/c9s/config.toml` pins `theme.name` are unaffected;
  to opt back in, delete the file or set `name = "dark"`.

### Fixed

- `:skin` now surfaces persistence failures in the toast (previously
  silent — only logged to `errors-*.log`). The persisted config is
  also `fsync`'d before close so the choice survives a Ctrl+C / sudden
  exit immediately after switching skins.

## [0.1.1] - 2026-05-03

This is a follow-up to v0.1.0 addressing the v0.1.0 code review. It
contains no new features. Two user-visible bugs are fixed (Shift+S sort
silently no-op'd; long sessions leaked goroutines/tickers), the
unimplemented plugin advertising is removed, the docs/CHANGELOG no
longer claim platforms or features that don't ship, and the release
pipeline + CI gain several robustness improvements.

### Fixed

- **Sort modal applied changes again** (review C1). Pressing Shift+S
  on every tabular screen used to open the column picker but the
  selection was silently discarded due to a value-vs-pointer receiver
  mismatch on the screens' ApplySort method. Standardized all screen
  constructors on `*Model` and pointer receivers; added a regression
  test that asserts the anonymous interface assertion succeeds for
  every sortable screen registered in the app.
- **Long-running c9s sessions no longer leak goroutines + tickers**
  (review C2). `clock.Real().Tick(d)` previously spawned a goroutine
  + `time.NewTicker(d)` per call that ran forever; with a 2-second
  refresh cadence this leaked ~1800 goroutines per screen per hour.
  Tick is now a thin wrapper around `time.After(d)` that the runtime
  GCs after firing.
- **`:skin` reload preserves screen state** (review I9). The skin
  command used to recreate every screen via its New() constructor,
  discarding filter, marks, sort key, and scroll position; five
  screens (errors, pinned, xray, pulses, jobs) weren't rebuilt at all
  and kept rendering with the old palette. Now broadcasts a
  `screens.PaletteChangedMsg` to every screen so colors refresh while
  internal state is preserved.
- **Race on `ProgressModel.awaitCancel`** (review I2). The 2-second
  cancel-window expiration used to write a Bubble Tea field from a
  `time.AfterFunc` callback, racing with the event loop. Replaced
  with `tea.Tick` returning `cancelWindowMsg{gen}` routed through
  Update; a generation counter guards against stale messages from
  prior windows.
- **Stream cancel watcher no longer leaks** (review M11). The
  goroutine watching for `ctx.Done()` blocked forever when the
  command exited naturally. Now coordinates with the reader via a
  `finished` channel: it exits immediately on natural completion and
  only escalates to SIGKILL after waiting up to 2 s on `finished`
  following SIGINT (review M7).
- **All UI→CLI calls now have a 5 s timeout** (review I6). 54 call
  sites under `internal/ui/screens` used `context.Background()`,
  which combined with C2 to compound goroutine accumulation behind a
  hung `container` subprocess. New `cli.DefaultCtx()` helper bounds
  every fetch/action.
- **`parsePruneCount` no longer mis-parses ID digits** (review M3).
  The old "first run of digits" heuristic would parse
  `Removed sha256:abc123def 5 containers` as 256. Replaced with a
  regex anchored on the unit word (containers/images/networks/volumes).
- **Pause/Unpause errors are detectable via errors.Is** (review M4).
  Introduced `cli.ErrUnsupported` sentinel; PauseContainer/
  UnpauseContainer wrap that instead of `context.Canceled`, so
  callers checking for cancellation no longer get false positives.
- **`pinned.List` now uses sort.Slice** (review M5).
- **`runVoid` preserves stdout in error wrappers** (review M12).
  Some `container` subcommands emit useful failure context on stdout;
  the wrapped error's Hint now falls back to stdout when stderr is
  empty.
- **Hardcoded `container` binary path eliminated from the shell-out
  path** (review M1). `cli.Client` now exposes `Bin()` returning the
  configured binary path; the interactive `s`-key shell-out uses it.
- **Docs site builds in `mkdocs --strict`** (review I3). Removed
  broken `docs/superpowers/...` links from `index.md`,
  `architecture.md`, `README.md`, and CHANGELOG; added a docs job to
  `ci.yml` that runs `mkdocs build --strict` to catch link rot
  before the Pages workflow does.
- **Stale `torosent/container-tui` references removed** (review I4).
  `mkdocs.yml`, `docs/quick-start.md`, `docs/faq.md`, and
  `docs/contributing.md` now point at `torosent/c9s`.
- **CHANGELOG corrected** (review I5). The v0.1.0 entry no longer
  claims darwin+linux × amd64+arm64 binaries (we only build
  darwin/arm64) and the duplicate `## [Unreleased]` heading is gone.

### Removed

- **Plugin system** (review I1). The v0.1.0 plugin loader was wired
  up but never invoked anywhere in `internal/ui` (verified by grep).
  Removed the `internal/plugins/` package, the `config.Plugin` type,
  the configuration docs, the FAQ entry, the k9s-migration row, and
  the README/CHANGELOG bullets that advertised it. Will reappear if
  and when the runtime hookup is implemented.

### Changed

- **`brews:` migrated to `homebrew_casks:`** in `.goreleaser.yaml`.
  GoReleaser deprecated `brews` in v2.10. The cask publishes to
  `Casks/c9s.rb` in the `torosent/homebrew-c9s` tap and includes a
  `postflight` xattr hook that strips the macOS quarantine attribute
  for unsigned binaries (replace with proper notarization once the
  Apple Developer account is wired up).
- **`replace_existing_artifacts: true`** in `.goreleaser.yaml`. Re-runs
  of the release workflow on the same tag now overwrite assets cleanly
  rather than failing with HTTP 422.
- **Goreleaser `before:` hook is `go mod verify`** instead of
  `go mod tidy` (review I7), so the release run can't mutate
  `go.mod`/`go.sum` away from what CI tested.
- **CI gains a "go mod tidy is a no-op" step** that fails if running
  `go mod tidy` would change the lock files.
- **Mkdocs nav** now includes architecture, install, keybinds,
  demos-rendering, and the previously orphaned screen pages
  (`build`, `errors`, `pinned`, `pulses`, `run`, `xray`).
- **Tap install instructions** simplified to the two-step form:
  ```
  brew tap torosent/c9s
  brew install c9s
  ```

### Internal

- Standardized screen constructors on returning `*Model`. All Update
  methods on screen types now use pointer receivers consistently.
- Added `cli.Fake.CallsCopy()` for thread-safe access to the recorded
  call list.
- Added `cli.Client.Bin()` to interface and both implementations.
- Added `internal/cli/timeout.go` with `DefaultCtx()` helper.
- Added a smoke + sortable-interface test for the previously
  zero-coverage `internal/ui/screens/pinned` package.
- Removed accidentally committed `internal/ui/screens/jobs/jobs.go.bak`.

## [0.1.0] - 2026-05-02

**This is the v0.1.0 cut — c9s's first proper release.** It collects:
- All the features shipped across v0.0.1 through v0.0.7 (foundations, containers screen, streaming/logs/jobs, all resource screens, configuration system, k9s parity, demos, docs site).
- A goreleaser-driven macOS Apple Silicon (darwin/arm64) binary.
- Homebrew tap at torosent/homebrew-c9s, distributed as a Homebrew Cask.
- GitHub Actions release workflow that publishes the binary + updates the cask.
- macOS codesigning + notarization hooks (optional; activate by setting required secrets).

### Added

**Release infrastructure (v0.1.0)**
- goreleaser configuration for darwin/arm64 build
- Homebrew tap at `torosent/homebrew-c9s` with auto-managed cask
- GitHub Actions release workflow triggered on version tags
- Optional macOS codesigning and notarization support
- Direct download archives with checksums
- Install documentation in README (Homebrew, direct download, build from source)
- Versioning policy documentation (0.x.y pre-1.0 semantics)

**Core features (v0.0.1 - v0.0.7)**
- **Terminal UI**: Fast, keyboard-driven interface inspired by k9s
- **Resource screens**: Containers, images, volumes, networks, builder, registries, DNS domains, system services, errors, pinned resources, pulses, xray diagnostics
- **Container operations**: Start, stop, restart, delete, inspect, streaming logs with color-coded multi-source output
- **Background jobs**: Job manager with progress tracking, cancellation, and dedicated jobs screen
- **Demo mode**: `--demo-data` flag populates fake client with sample data for exploration without a runtime
- **Themes**: Customizable color schemes (k9s skin format supported) with 8 built-in themes
- **Mouse support**: Click row select, wheel scroll, drag support
- **Keyboard navigation**: vim-style bindings, colon commands (`:quit`, `:jobs`, etc.), sortable columns (Shift+S)
- **Configuration**: `~/.config/c9s/config.yaml` for theme, refresh intervals, and keybindings
- **Plugin system**: Extensible command execution and screen injection
- **Logs**: Multi-source streaming with ANSI color preservation and buffering
- **Metrics**: Live CPU/memory/disk stats for containers
- **Documentation**: Auto-generated hotkeys reference, mkdocs-material site with quick-start and FAQ
- **Demos**: VHS tape scripts and animated GIFs for key workflows

### Changed

- Sortable + mouse support extended to all tabular screens
- Breadcrumb trail wired into root model for navigation history
- Test coverage maintained at 74.4% (above 70% quality gate)

### Fixed

- Fixed `gen-hotkeys` tool to use `titleCase` instead of deprecated `strings.Title`

## [0.0.7] - 2026-05-02

### Added

- **Demo mode**: `--demo-data` flag populates fake client with sample containers, images, volumes, networks, running builder, registries, DNS domains, and system services (T1)
- **Jobs screen**: Integrated into screen registry with sorting, filtering, and mouse support; access via `:jobs` command (T2)
- **Hotkeys documentation**: Auto-generated `docs/hotkeys.md` with global and per-screen keyboard reference; regenerate with `make docs-hotkeys` (T3)
- **Documentation site**: mkdocs-material setup with quick-start guide, FAQ, contributing guide, and navigation structure (T4)
- **Demo scripts**: VHS tape files in `tools/demos/` for containers, logs, builder, and jobs demos (T5)
- **Animated GIF**: README now features `docs/assets/containers.gif` placeholder and enhanced feature list (T6)
- **CI/CD**: GitHub Pages workflow deploys docs; demos workflow renders VHS tapes and commits GIFs automatically (T7)

### Changed

- Jobs screen refactored to implement `screens.Screen` and `screens.Sortable` interfaces
- Coverage maintained at 74.3% (above 70% quality gate)

### Fixed

- Fixed `gen-hotkeys` tool to use `titleCase` instead of deprecated `strings.Title`

## [0.0.6] - 2025-05-02

### Added
- Sort UI: Shift+S opens column picker modal for sortable screens (containers)
- Breadcrumb trail package with Push/Pop/Render (foundation ready for wiring)
- Mouse support: click row select, wheel scroll via tea.WithMouseCellMotion
- Coverage backfill: pulses (85%), xray (80%), widgets (86%)
- Test coverage for breadcrumbs, sortpicker modal, mouse handling
- Sortable interface for screens with per-column sorting

### Changed
- Container screen now supports sorting by id/image/status/uptime/cpu/mem
- Total test coverage: 77.6%

## [0.0.5] - 2026-05-02

### Added
- Layered configuration system with TOML-based config files
- `config.toml` for general settings (readonly, refresh intervals, theme, stream)
- `hotkeys.toml` for keybind overrides (supports remapping any action)
- `aliases.toml` for palette command aliases (e.g., `kpods = "containers"`)
- `views.toml` for per-screen column visibility, order, and width customization
- Live config reload with fsnotify - changes apply without restart
- Read-only mode (`--readonly` flag or `[ui] readonly = true` in config)
- Destructive commands blocked in read-only mode with toast notification
- `[READONLY]` indicator in status bar when read-only mode active
- Configuration precedence: CLI flags > config.toml > built-in defaults
- Config loading from `$XDG_CONFIG_HOME/c9s/` or `~/.config/c9s/`
- Skin system with 4 bundled themes: dark, light, k9s-dark, k9s-light
- `:skin <name>` palette command to switch themes at runtime
- k9s skin importer: `c9s import-skin <k9s-skin.yaml>` and `:import-skin` command
- User skins loadable from `~/.config/c9s/skins/<name>.toml`
- Plugin loader for user-defined commands from `~/.config/c9s/plugins/*.toml`
- Plugin TOML format with scope, key, command template, and variable substitution
- Build EWMA persistence to `~/.local/share/c9s/build-stats.toml` for time estimates
- Documentation: `docs/configuration.md` with comprehensive examples
- Documentation: `docs/skins.md` for skin system and k9s compatibility

## [0.0.4] - 2026-05-02

### Added
- `:images` resource screen with REPO·TAG·ID·CREATED·SIZE table, 2 s refresh,
  marks, and hotkeys (`d` inspect, `t` tag via TextInput, `P` push via progress
  modal, `D` delete with confirm, `R` run-from-image opens the run form).
- `:volumes` resource screen with NAME·DRIVER·MOUNTPOINT·SIZE·USED-BY columns
  and `d`/`D` row hotkeys.
- `:networks` resource screen with NAME·DRIVER·SUBNET·CONTAINERS columns and
  `d`/`D` row hotkeys.
- `:builder` single-card screen with STATE·CPU·MEM·UPTIME, color-coded state,
  and `S`/`X`/`D`/`r` keys.
- `:registry` screen with HOST·USER·DEFAULT, `L` opens login modal, `D`
  logs out (confirm), `*` sets default. Login modal masks the password and
  pipes it via stdin (never argv) to `container registry login --password-stdin`.
- `:system` services table (SERVICE·STATE·PID·UPTIME) with `S`/`X` start/stop-all.
- System sub-screens: `:df` (read-only), `:dns` (CRUD with `c`/`D`/`*`),
  `:property` (edit `e`, reset `D`; read-only properties protected with toast),
  `:kernel` (read-only viewport for `kernel.*` properties),
  `:logs` (streaming `container system logs --follow` with ring buffer + auto-follow).
- Multi-field form modals: `RunFormModel` (`:run`, also opened by `R` on Images;
  pre-fillable image), `BuildFormModel` (`:build`). Tab cycles fields, Space
  toggles booleans, Ctrl-Enter / Ctrl-D submits, Esc cancels.
- Generic `TextInput` modal (`modals.NewTextInput(label, prompt, initial, palette)`)
  with optional validator. Used by `:create`, `:tag`, `:save`, `:load`, and the
  per-screen prompts (DNS create, property edit, image tag).
- Palette commands: `:images`, `:i`, `:volumes`, `:v`, `:networks`, `:n`,
  `:builder`, `:b`, `:registry`, `:reg`, `:system`, `:sys`, `:df`, `:dns`,
  `:property`, `:kernel`, `:logs`, `:run [image]`, `:build [path]`,
  `:login [host]`, `:create`, `:tag`, `:save`, `:load`, `:pull <ref>`,
  `:push <ref>`.
- Per-screen docs: `docs/screens/{images,volumes,networks,builder,registry,system,run,build}.md`.
  `docs/keybinds.md` updated with the full palette + per-screen index.

### Changed
- `cli.Client` interface gained 24 new methods covering image (`ListImages`,
  `InspectImage`, `TagImage`, `DeleteImage`, `PruneImages`, `LoadImage`,
  `SaveImage`), volume, network, builder, registry, system, and run (including
  `StreamSystemLogs` and `RunContainer` returning a `Stream`).
- `cli.RegistryLogin` sends the password via the child process's stdin, never
  via argv.
- Tolerant JSON parsers: empty / `null` / unrecognized output now returns an
  empty slice (or zero-value struct) so screens render gracefully instead of
  surfacing a parser error.

## [0.0.3] - 2026-05-02

### Added
- Streaming `Stream` type in `internal/cli` with build/pull/push/logs parsers.
- `internal/jobs.Manager` for tracking concurrent streaming jobs (thread-safe, race-tested).
- Log viewer modal with ring buffer (5000 lines), filter (`/`), follow-tail (`G`), save (`Ctrl+S`), level coloring (INFO=cyan, WARN=yellow, ERROR=red, DEBUG=dim), `t`/`T` timestamp toggles, multi-source merge with `[name]` prefix coloring.
- Progress modal with build (step list + raw viewport, `v` toggle) and pull/push (layer table with mini progress bars) variants. `Ctrl+C` double-tap to cancel (2s window), `Ctrl+Z` detaches to `:jobs`.
- `:jobs` screen lists running/completed background jobs with columns ID/KIND/TARGET/STATE/ELAPSED/LINES. `Enter` re-attaches modal, `Ctrl+C` cancels job, `D` clears done.
- Palette commands: `:logs <id>`, `:build <path>`, `:pull <ref>`, `:push <ref>`, `:run <image>`, `:jobs`.
- `l` row hotkey on containers screen opens log viewer (single or multi-source if marks set).
- `docs/streaming.md` comprehensive guide.

### Changed
- `cli.Client` interface extended with `StreamLogs`, `StreamBuild`, `StreamPull`, `StreamPush` methods.
- `theme.SourceColors` palette for multi-source coloring.

## [0.0.2] - 2026-05-02

### Added
- `:containers` resource screen with table, 2-second refresh, marks, and hotkeys (s shell, d inspect, x stop, Shift+K for removal operations, Shift+R restart, Shift+D delete, p pause).
- Confirm/Help/Inspect modals with a stack-based modal system.
- `Screen` interface + screen registry in the root model. Other resource screens are registered as "not yet available" placeholders for Plan 4.
- `keymap.Map` with default + override merge.
- `tea.ExecProcess` suspend-and-shell-out for `s` (container exec).
- Capability gating: `p` (pause) emits a status toast when unsupported and is annotated in help.

### Changed
- Root model refactored from a flat splash + placeholder to a router that delegates to active screen + modal stack.
- Toolchain pin bumped to Go 1.24.2 to match newer transitive dependencies.

### Fixed
- `InspectModel.View` no longer ineffectively assigns to value-receiver state; viewport is now initialised in the constructor and resized per render.

## [0.0.1] - 2026-05-02

### Added
- Project skeleton: Go module, Bubble Tea root, splash, status bar, capabilities probe.
- CLI gateway interface (`internal/cli`) with `DefaultClient` (shells out to `container`) and `Fake` for tests.
- Clock abstraction (`internal/clock`) and snapshot cache (`internal/state`).
- Continuous integration workflow with lint, race-detector tests, and coverage gate.
- Architecture, install, and index pages (placeholders).
