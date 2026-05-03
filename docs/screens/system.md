# System screens (`:system`, `:sys`)

`:system` lands on the **services** table; sub-screens are reachable via the
command palette.

## `:system` (services)

Columns: **SERVICE · STATE · PID · UPTIME**. Refreshes every 2 seconds.

| Key | Action |
|---|---|
| `S` | Start all system services |
| `X` | Stop all system services |
| `r` | Refresh |
| `/` | Filter by name or state |

## `:df` — disk usage

Read-only summary of `container system df` rendered as three rows
(Images / Containers / Volumes), each showing **COUNT · ACTIVE · SIZE ·
RECLAIMABLE**. The status bar Summary shows total + reclaimable.

| Key | Action |
|---|---|
| `r` | Refresh |

## `:dns` — local DNS domains

Columns: **NAME · DEFAULT** (`*` marks the default).

| Key | Action |
|---|---|
| `c` | Create domain (TextInput prompt) |
| `D` | Delete (confirm) |
| `*` | Set focused domain as default |
| `/` | Filter |
| `r` | Refresh |

## `:property` — system properties

Columns: **KEY · VALUE · RO** (`*` marks read-only).

| Key | Action |
|---|---|
| `e` | Edit value (TextInput prompt) |
| `D` | Reset to default (confirm) |
| `/` | Filter |
| `r` | Refresh |

Read-only properties surface a toast instead of opening the edit/reset modal.

## `:kernel` — kernel configuration

Read-only viewport listing all `kernel.*` system properties as
`key = value` lines. `r` refreshes.

## `:logs` — system log stream

Streaming viewer wired to `container system logs --follow`. Maintains a
5000-line ring buffer; `g`/`G` jump to top/tail; `j`/`k` scroll; auto-follows
new lines unless the user has scrolled up. Closing the screen cancels the
stream.
