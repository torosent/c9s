# Streaming + Logs + Jobs

c9s v0.0.3 adds streaming output support for build/pull/push operations, a log viewer modal for container logs, and a jobs panel for managing background streams.

## Log Viewer

Open with:
- `:logs <id>` from anywhere
- `l` on a container row (single-source)
- `l` with marked containers (multi-source merge)

### Features

**Ring Buffer**: Holds the last 5000 lines. Older lines are automatically dropped.

**Follow-Tail**: Enabled by default. Viewport automatically scrolls to bottom as new lines arrive. Disabled when you scroll up; re-enable with `G`.

**Filter**: Press `/` to enter filter mode. Type a substring and press Enter. Only matching lines are shown. Press `/` again and clear input to remove filter.

**Save**: `Ctrl+S` saves the current buffer to `~/.local/share/c9s/logs/<name>-<unix>.log`.

**Timestamps**:
- `t` toggles wall-clock timestamp column
- `T` toggles relative-time column ("3s ago")

**Level Coloring**: Lines starting with `INFO` / `WARN` / `ERROR` / `DEBUG` are automatically colored (cyan / yellow / red / dim).

**Multi-Source**: When multiple containers are selected, each line is prefixed with `[container-name]` in a stable color.

### Hotkeys

| Key | Action |
|-----|--------|
| `q`, `Esc` | Close |
| `/` | Enter filter mode |
| `G` | Jump to bottom (re-enable follow-tail) |
| `g` | Jump to top |
| `t` | Toggle wall-clock timestamp |
| `T` | Toggle relative timestamp |
| `Ctrl+S` | Save buffer to disk |
| `↑`, `k` | Scroll up one line |
| `↓`, `j` | Scroll down one line |
| `PgUp` | Scroll up one page |
| `PgDown` | Scroll down one page |

---

## Progress Modal

Streams output from `container build`, `container image pull`, and `container image push`.

### Build View

**Step List** (default): Shows BuildKit steps with icons:
- `○` waiting
- `●` running
- `✓` done
- `⚡` cached

Press `v` to toggle to raw output view (full stdout/stderr).

**Double-Tap Cancel**: Press `Ctrl+C` once to enter cancel mode (footer shows warning). Press `Ctrl+C` again within 2s to send SIGINT.

**Detach**: Press `Ctrl+Z` to detach the job into `:jobs` (modal closes, stream continues in background).

### Pull/Push View

Shows a layer table with mini progress bars:

```
sha256:abc123...  downloading  ██████████░░░░░░░░░░ 50%
sha256:def456...  extracting   ████████████████████ 100%
sha256:ghi789...  mounted
```

Aggregate header shows: `pulling <ref> · X/Y layers · A MB / B MB · Z MB/s`.

Same `Ctrl+C` (double-tap) and `Ctrl+Z` (detach) behavior as build.

### Hotkeys

| Key | Action |
|-----|--------|
| `q`, `Esc` | Close (when done) |
| `Ctrl+C` | Cancel (double-tap) |
| `Ctrl+Z` | Detach to :jobs |
| `v` | Toggle step list ↔ raw output (build only) |

---

## Jobs Panel

Access with `:jobs` from anywhere.

### Columns

- **ID**: Job identifier
- **KIND**: `build`, `pull`, `push`, `log`
- **TARGET**: Image ref, container ID, or path
- **STATE**: `running`, `done`, `failed`, `cancelled`
- **ELAPSED**: Duration since start
- **LINES**: Count of accumulated output lines

Refreshes every 1 second.

### Hotkeys

| Key | Action |
|-----|--------|
| `Enter` | Re-attach to job (opens appropriate modal) |
| `Ctrl+C` | Cancel selected job |
| `D` | Clear done/failed jobs |
| `q` | Return to previous screen |

---

## Palette Commands

All streaming operations are accessible via the command palette (`:`).

| Command | Description |
|---------|-------------|
| `:logs <id>` | Stream logs from container |
| `:build <path>` | Build image from path (opens progress modal) |
| `:pull <ref>` | Pull image (opens progress modal) |
| `:push <ref>` | Push image (opens progress modal) |
| `:run <image>` | Quick-run container (streams output) |
| `:jobs` | Switch to jobs panel |

---

## Implementation Notes

**Stream Type**: `cli.Stream` wraps `Events <-chan StreamEvent` and `Done <-chan StreamResult`. Context cancellation sends SIGINT, then SIGKILL after 2s.

**Job Manager**: `jobs.Manager` tracks background streams. Thread-safe with RWMutex. List/Get return copies to prevent data races.

**Parsers**:
- `parsers_logs.go`: Regex detects `INFO|WARN|ERROR|DEBUG` at line start.
- `parsers_build.go`: Tolerant BuildKit step parser; emits `RawLine` for unrecognized output.
- `parsers_layers.go`: State machine maintains `digest → LayerProgress` map.

All parsers tested with shell script fixtures.

---

## Examples

**Single-source logs**:
```
:logs api-server
```

**Multi-source logs** (mark containers first):
```
space   (mark api-server)
space   (mark worker-1)
l       (open merged log viewer)
```

**Build with detach**:
```
:build /path/to/project
(wait for first few steps)
Ctrl+Z  (detach to background)
:jobs   (see it running)
```

**Re-attach to job**:
```
:jobs
Enter   (on the row)
```

---

For full feature list, see `docs/screens/containers.md` and `docs/keybinds.md`.
