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
- `:acr-login <registry>` — one-step Azure Container Registry login via
  Azure AD. Accepts either `myregistry` or `myregistry.azurecr.io`.
  Requires the [Azure CLI](https://learn.microsoft.com/cli/azure/install-azure-cli)
  on `PATH` and a current `az login` session. See [ACR login](#azure-container-registry-acr).

## Azure Container Registry (ACR)

Apple's `container registry login` accepts any OCI v2 registry, so ACR
works with the standard login modal once you have credentials. The
`:acr-login` palette command automates the most common case.

| ACR auth method | What to type in `:login` |
|---|---|
| **Admin user** (must be enabled on the registry) | username = registry name, password = admin key |
| **Service principal** (Entra ID) | username = SP appId, password = SP secret |
| **AAD access token** (recommended for humans) | username = `00000000-0000-0000-0000-000000000000`, password = output of `az acr login --name <acr> --expose-token --output tsv --query accessToken` |

### `:acr-login` (one-step AAD flow)

The `:acr-login <registry>` command wraps the AAD-token recipe end-to-end:

1. Runs `az acr login --name <registry> --expose-token` (you must have
   already authenticated `az` with `az login`).
2. Calls `container registry login <registry>.azurecr.io --username
   00000000-0000-0000-0000-000000000000 --password-stdin` with the
   resulting token, so it never appears on argv.
3. Surfaces success or failure as a toast.

The token is short-lived (~3 hours). Re-run `:acr-login` when it
expires; pulls and pushes meanwhile use the persisted credential the
same way they do for any other registry.

### Sovereign clouds

ACR hostnames in sovereign clouds (`*.azurecr.us`, `*.azurecr.cn`) are
preserved as-is when passed to `:acr-login`, so
`:acr-login myreg.azurecr.us` does the right thing.
