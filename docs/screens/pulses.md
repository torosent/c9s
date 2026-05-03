# Pulses Screen

The **pulses** screen is a live dashboard showing real-time system metrics.

## Access

- Command: `:pulses`

## Features

- **Auto-refresh**: Updates every 2 seconds
- **Resource counts**: Displays running containers, images, volumes, networks
- **System status**: Shows current time and health status
- **Card layout**: Organized information in visual cards

## Dashboard Cards

### Resources Card
- Containers: Count of running containers
- Images: Total images available
- Volumes: Total volumes
- Networks: Total networks

### System Card
- Current time (HH:MM:SS)
- System status (Healthy/Warning/Error)

## Example

```
┌─────────────────────┐  ┌─────────────────────┐
│ Resources           │  │ System              │
│                     │  │                     │
│ Containers: 12      │  │ Time: 14:32:45      │
│ Images:     47      │  │ Status: Healthy     │
│ Volumes:    8       │  │                     │
│ Networks:   3       │  │                     │
└─────────────────────┘  └─────────────────────┘
```

## Hotkeys

| Key | Action |
|-----|--------|
| `r` | Manual refresh |
| `Esc` or `q` | Close screen |

## Use Cases

- Quick overview of system resources
- Monitor container counts during deployments
- Verify system health at a glance
- Dashboard for monitoring setups

## Implementation Notes

- Lightweight polling (no expensive operations)
- Uses `ListContainers(false)` for running only
- Future: Add CPU/memory metrics, alert indicators
