# Docker compatibility shim

c9s ships a small POSIX shell shim that maps the most common
`docker(1)` subcommands to their Apple `container` CLI equivalents.
With the shim on `PATH` (and shadowing any pre-existing docker
binary), familiar invocations like `docker ps`, `docker images`, or
`docker pull <ref>` "just work" against Apple Containers without
relearning the verb shape.

This is a best-effort compatibility shim, not a drop-in replacement —
docker has many flags Apple's CLI doesn't expose and vice versa. The
shim translates **command shape** only and forwards remaining argv
as-is.

## Install

From a shell:

```bash
c9s install-docker-shim
```

By default it writes `~/.local/bin/docker`. Override with `--path`:

```bash
c9s install-docker-shim --path /usr/local/bin/docker
c9s install-docker-shim --path /usr/local/bin/docker --force   # overwrite existing
```

If a non-shim file already lives at the target path, install refuses
unless `--force` is passed. This protects an existing real docker
binary from being clobbered silently.

You can also install from inside the TUI via the palette:

```
:install-docker-shim                # ~/.local/bin/docker
:install-docker-shim /custom/path   # any path
```

After installing, make sure the install directory is on `PATH` and
appears **before** any directory containing a real docker:

```bash
echo $PATH | tr ':' '\n'
which docker          # should show your shim, not /Applications/Docker.app/...
```

If your shell caches command lookups, run `hash -r` (bash, zsh) or
`rehash` (zsh).

## Uninstall

```bash
c9s uninstall-docker-shim                  # default path
c9s uninstall-docker-shim --path /elsewhere
# or in the TUI:
:uninstall-docker-shim
```

Uninstall checks the file for the c9s sentinel header and refuses to
delete a file that doesn't carry it. If you have a real docker binary
at the target path, uninstall will not remove it.

## What it translates

| You type | The shim runs |
|---|---|
| `docker run` / `start` / `stop` / `restart` / `kill` / `exec` / `inspect` / `logs` / `build` / `create` / `delete` | `container <verb>` (passthrough) |
| `docker rm <id>` | `container delete <id>` |
| `docker ps` / `docker ps -a` | `container list` / `container list --all` |
| `docker images` | `container image list` |
| `docker rmi <id>` | `container image delete <id>` |
| `docker pull <ref>` / `push` / `tag` | `container image <verb>` (all three live in Apple's `image` namespace) |
| `docker volume ls` / `docker network ls` | `container volume list` / `container network list` |
| `docker volume <other>` / `docker network <other>` | `container volume <other>` / `container network <other>` |
| bare `docker volume` / `docker network` | `container volume` / `container network` (Apple's CLI shows its own help) |
| `docker login` / `docker logout` | `container registry login` / `container registry logout` |
| `docker version` | `container --version` |
| `docker info` | `container system info` |
| `docker system <args>` | `container system <args>` |
| anything else | `container <verb> <args>` (fall-through) |

The shim respects a `CONTAINER_BIN` environment variable so users with
the `container` binary at a non-standard path can override the lookup.

## When to use it

- You have shell scripts or muscle memory built around `docker` and
  don't want to rewrite them.
- You're following a tutorial that uses `docker` and want to follow
  along without translating commands in your head.

## When **not** to use it

- You depend on docker-only behavior (`docker compose`, BuildKit
  cache mounts, build secrets via the docker CLI's syntax, etc.).
  None of those are translated; you'll need to use Apple's CLI
  natively or fall back to a real docker.
- You need pinned exit codes / output formatting compatible with
  docker — the shim doesn't massage stdout/stderr.

## Detection of existing Docker

Before writing the shim, `c9s install-docker-shim` walks `PATH` and lists
any other `docker` executable it finds (in PATH order, so the *first*
entry is the one shells will currently resolve `docker` to). It also
checks for Docker Desktop at `/Applications/Docker.app`. Both checks
print warnings to stderr but don't block install — they're informational
so you can decide whether to reorder PATH afterwards.

Example output when an existing docker is on PATH:

```
$ c9s install-docker-shim
Warning: an existing docker binary is already on PATH:
  - /usr/local/bin/docker
After install, make sure the shim's directory is BEFORE
the directory containing the existing docker on PATH, or
the existing binary will continue to win.

Warning: Docker Desktop is installed at /Applications/Docker.app.
While Docker Desktop is running, the system docker daemon will
still respond to whichever 'docker' binary your shell resolves
first. The shim only redirects CLI calls to Apple containers; it
doesn't affect Docker Desktop's running containers either way.

Installed c9s docker shim to: /Users/you/.local/bin/docker
...
```

Inside the TUI, the same checks compress into a single toast — the
existing docker path and/or "Docker Desktop detected" tag is appended
to the install confirmation.

## Limitations

- `docker compose` is not handled. Run Apple's `container` directly
  for orchestration.
- Flag remapping is **not** done — only verb remapping. So e.g.
  `docker ps --format '{{.Names}}'` falls through unchanged and
  Apple's CLI may reject the format string. Same goes for `docker
  logs -f`: argv passes through unchanged, so it works only if
  Apple's CLI accepts `-f` as a synonym (or you use `--follow`
  directly).
- The script is regenerated by `c9s install-docker-shim`, so if Apple's
  command shape changes you can re-run the subcommand to pick up an
  updated mapping table.
