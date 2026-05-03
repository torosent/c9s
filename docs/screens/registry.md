# Registry screen (`:registry`, `:reg`)

Lists configured registry logins, refreshing every 2 seconds.

## Columns

- **HOST** — registry host (e.g. `ghcr.io`).
- **USER** — username for the saved credential.
- **DEFAULT** — `*` for the default registry.

## Hotkeys

| Key | Action |
|---|---|
| `↑/↓` or `j/k` | Move selection |
| `space` | Mark focused row |
| `*` | Set focused registry as default |
| `Esc` | Clear marks |
| `/` | Filter by host or user |
| `r` | Refresh |
| `L` | Open the login modal |
| `D` | Logout (confirm) |

## Login flow

The login modal is a 3-field form: **Host**, **User**, **Password**.
The password field uses the masked echo mode (`*` placeholder rendered for
each character) so the password is never echoed to the terminal nor stored
in any view buffer.

When the user submits, c9s calls `container registry login <host>
--username <u> --password-stdin` and pipes the password on **stdin**. The
password never appears in the process argv list.

## Palette commands

- `:login <host>` — open the login modal pre-filled with `<host>`.
