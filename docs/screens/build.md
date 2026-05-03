# Build modal (`:build [path]`)

The build modal wraps `container build`.

## Fields

| Field | Notes |
|---|---|
| **Path** | Build context directory. Pre-filled by `:build <path>`. |
| **Tag** | `-t <ref>` |
| **Containerfile** | `-f <path>`. Defaults to `Containerfile` if empty. |
| **Platform** | `--platform <plat>`. Empty leaves the runtime default. |

## Keymap

Identical to the [run modal](run.md): Tab/Shift-Tab cycle, Enter advances,
Ctrl-Enter submits, Esc cancels.

## Submission

c9s calls `cli.StreamBuild` and opens the build progress modal (rolling
step list with synthetic progress bars, raw-output toggle on `v`,
detach with `Ctrl-Z`, cancel with `Ctrl-C` twice within 2 s).
