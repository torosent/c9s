# Run modal (`:run [image]`, or `R` from the Images screen)

The run modal is a multi-field form that wraps `container run`.

## Fields

| Field | Notes |
|---|---|
| **Name** | Optional. If empty, the runtime auto-suggests one. |
| **Image** | Required. Pre-filled when invoked from the Images screen or `:run <image>`. |
| **Ports** | Comma-separated list of `host:container` mappings (e.g. `8080:80,443:443`). Each becomes one `-p` flag. |
| **Env** | Comma-separated `KEY=val` pairs (e.g. `KEY=val,OTHER=2`). Each becomes one `-e` flag. |
| **Volumes** | Comma-separated `src:dst` mounts. Each becomes one `-v` flag. |
| **Interactive** | Toggle (`-i`). |
| **TTY** | Toggle (`-t`). |
| **Detach** | Toggle (`-d`). **Default on.** |

## Keymap

- `Tab` / `Shift-Tab` — cycle fields.
- `Space` — toggle the focused boolean field.
- `Enter` — advance to next field; on the last field, submit.
- `Ctrl-Enter` (or `Ctrl-D`, `Ctrl-S` for terminals lacking Ctrl-Enter) — submit immediately.
- `Esc` — cancel.

## Submission

On submit, c9s opens a progress modal streaming `container run` output. On
success the modal exits and (Plan 5) c9s switches to `:containers` and
highlights the new row.
