# Rendering demo GIFs locally

The `.tape` files under `tools/demos/` describe screen recordings of c9s
flows for the README and docs site. They are rendered with
[VHS](https://github.com/charmbracelet/vhs).

## Prerequisites

```bash
brew install vhs ttyd
make build
```

## Render all tapes

```bash
mkdir -p docs/assets
for tape in tools/demos/*.tape; do
    echo "Rendering $tape..."
    vhs "$tape"
done
```

## Render a specific tape

```bash
vhs tools/demos/containers.tape
```

The output GIF is written to `docs/assets/<name>.gif` per the `Output`
directive in each tape file.

## Why no CI workflow?

A previous version of this repo had `.github/workflows/demos.yml` to
auto-render on a schedule. It was removed because rendering reliably in CI
requires `ttyd` + `vhs` + a valid binary execution path that is sensitive
to ubuntu-latest image changes. Render locally and commit the resulting
GIFs; that's faster and more reliable.

If you want to re-add the workflow, see the [VHS GitHub action
docs](https://github.com/charmbracelet/vhs#vhs-github-actions).
