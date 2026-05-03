# Architecture

`c9s` is a single Go binary built on the [Bubble Tea](https://github.com/charmbracelet/bubbletea) Elm-architecture TUI framework. It manages Apple's `container` runtime by shelling out to the `container` CLI; it never talks to the underlying XPC services directly.

```
ui ─→ state ─→ cli ─→ `container` (subprocess)
```

- **`internal/cli`** is the only package that runs `container`. Production code uses `cli.DefaultClient`; tests use `cli.Fake`.
- **`internal/state`** holds per-resource snapshot caches with freshness timestamps.
- **`internal/ui`** owns the Bubble Tea root model, modal stack, splash, status bar, and individual resource screens (containers, images, volumes, networks, builder, registry, system).

For the full design, see the [design spec](superpowers/specs/2026-05-02-c9s-design.md).
