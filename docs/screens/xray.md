# XRay Screen

The **xray** screen displays resource relationships as an interactive tree view.

## Access

- Command: `:xray` or `:x`

## Features

- **Tree hierarchy**: Shows containers with their parent images as children
- **Expand/collapse**: Use `→/←` or `Space` to expand/collapse nodes
- **Navigation**: Use `↑/↓` or `j/k` to move through the tree
- **Jump to resource**: Press `Enter` to navigate to the resource's screen

## Tree Structure

```
Containers
  ▶ nginx:latest (abc123)
    └─ Image: nginx:latest
  ▼ redis:7-alpine (def456)
    └─ Image: redis:7-alpine
  ▶ postgres:14 (ghi789)
    └─ Image: postgres:14
```

## Hotkeys

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate tree nodes |
| `→` or `Space` | Expand node |
| `←` | Collapse node |
| `Enter` | Jump to resource |
| `Shift+E` | Expand all nodes |
| `Shift+C` | Collapse all nodes |
| `Esc` or `q` | Close screen |

## Use Cases

- Understand which containers use which images
- Identify orphaned images (no running containers)
- Navigate complex multi-container setups
- Visualize resource dependencies

## Implementation Notes

- Custom tree widget (Bubble Tea doesn't provide one)
- Parent links for efficient navigation
- Lazy loading for large container sets
