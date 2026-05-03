# Errors Screen

The **errors** screen displays a log of all errors that occurred during c9s operations.

## Access

- Command: `:errors` or `:e`
- Automatically populated when operations fail (build, run, pull, push, skin changes, etc.)

## Features

- **JSONL log file**: Errors are persisted to `~/.local/share/c9s/errors-YYYY-MM-DD.log`
- **Daily rotation**: New log file created each day
- **Table view**: Shows timestamp, operation, resource, and message
- **Detail modal**: Press `Enter` on any error to see full details
- **Copy to clipboard**: Press `y` to copy selected error as markdown

## Error Structure

Each error entry contains:
- `time`: ISO 8601 timestamp
- `op`: Operation that failed (e.g., "run", "build", "pull")
- `resource`: Resource ID or name involved
- `message`: Short error message
- `detail`: Full error details (stack trace, stderr output, etc.)

## Hotkeys

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate entries |
| `Enter` | View error details |
| `y` | Copy error as markdown |
| `Esc` or `q` | Close screen |

## Example

```
┌─ Errors ───────────────────────────────────────┐
│ Time       Operation  Resource     Message     │
├────────────────────────────────────────────────┤
│ 14:32:01   run        nginx:latest ENOENT      │
│ 14:15:22   build      myapp:v2     Go compile  │
│ 13:45:10   pull       redis:7      Network     │
└────────────────────────────────────────────────┘
```
