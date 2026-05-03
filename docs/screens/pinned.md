# Pinned Screen

The **pinned** screen shows bookmarked resources for quick access.

## Access

- Command: `:pinned` or `:p`
- Bookmark any resource by pressing `b` on containers/images/volumes/networks screens

## Features

- **Persistent storage**: Bookmarks saved to `~/.local/state/c9s/pinned.toml`
- **Jump to resource**: Press `Enter` to navigate to the resource's screen
- **Unpin**: Press `D` to remove bookmark
- **Display name**: Shows friendly name for each resource

## Hotkeys

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate bookmarks |
| `Enter` | Jump to resource screen |
| `D` | Unpin/remove bookmark |
| `Esc` or `q` | Close screen |

## Bookmark Structure

```toml
[[pins]]
resource = "container"
id = "abc123"
display = "nginx:latest (abc)"
added = "2026-05-02T14:30:00Z"
```

## Example

```
┌─ Pinned Bookmarks ─────────────────────────────┐
│ Resource   ID       Display                    │
├────────────────────────────────────────────────┤
│ container  abc123   nginx:latest (abc)         │
│ image      sha256…  redis:7-alpine             │
│ volume     data_vol myapp-data                 │
└────────────────────────────────────────────────┘
```

## Use Cases

- Pin production containers for quick monitoring
- Bookmark frequently inspected images
- Mark critical volumes for easy access
