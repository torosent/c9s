# c9s — comprehensive demo against the demo dataset.
#
# Single dark-theme walk-through covering every major scenario:
#   1. Splash → containers screen
#   2. Help overlay (?)
#   3. Filter (/) and sort (Shift+S)
#   4. Inspect, pin, prune-stopped (Shift+P) on the containers screen
#   5. Number-shortcut navigation through every resource screen
#      (1 containers · 2 images · 3 volumes · 4 networks · 5 builder
#       · 6 registry · 7 system · 8 pulses · 9 xray · 0 pinned)
#   6. System sub-screens (:df, :dns, :property, :kernel)
#   7. Action modals: run-from-image, build, registry login
#   8. New v0.1.x features: :acr-login, :install-docker-shim
#   9. :errors, :jobs
#  10. Quit
#
# Render locally with:
#   brew install vhs ttyd   # one-time
#   make build              # produces ./bin/c9s
#   vhs tools/demos/c9s.tape
#
# The default c9s theme is now k9s-gruvbox-dark; we set the terminal
# background to match so the box border doesn't sit on a contrasting
# strip of background.

Output docs/assets/c9s.gif

Set Theme "GruvboxDark"
Set Width 1400
Set Height 800
Set FontSize 14
Set Padding 20
Set TypingSpeed 35ms

# ── Launch ──────────────────────────────────────────────────────────
# Use a fresh XDG_CONFIG_HOME so c9s picks up the v0.1.4 default skin
# (k9s-gruvbox-dark) regardless of whatever the recording user has
# persisted locally. The setup is hidden from the GIF.
Hide
Type "rm -rf /tmp/c9s-demo-config && mkdir -p /tmp/c9s-demo-config && export XDG_CONFIG_HOME=/tmp/c9s-demo-config && clear"
Enter
Sleep 200ms
Show

Type "./bin/c9s --demo-data"
Sleep 400ms
Enter
Sleep 1500ms

# Splash dismissal
Sleep 1500ms
Space
Sleep 1500ms

# ── Containers screen tour ──────────────────────────────────────────
# Help overlay
Type "?"
Sleep 2200ms
Escape
Sleep 800ms

# Filter
Type "/"
Sleep 300ms
Type "api"
Sleep 1200ms
Escape
Sleep 800ms

# Sort picker (Shift+S, then 'r' to sort by something distinctive,
# Esc to dismiss without locking in)
Type "S"
Sleep 1200ms
Down
Sleep 300ms
Down
Sleep 300ms
Escape
Sleep 800ms

# Inspect focused container
Type "d"
Sleep 1500ms
Escape
Sleep 700ms

# Pin focused container (b) so :pinned later has a row
Type "b"
Sleep 900ms

# Prune stopped containers (Shift+P) — open the confirm modal,
# then cancel so the demo dataset stays intact
Type "P"
Sleep 2000ms
Type "n"
Sleep 800ms

# ── Resource screens via number shortcuts ───────────────────────────
Type "2"
Sleep 1300ms

# Run modal from images: focus a row, hit R (run-from-image)
Down
Sleep 300ms
Type "R"
Sleep 1800ms
Escape
Sleep 700ms

Type "3"
Sleep 1200ms
Type "4"
Sleep 1200ms
Type "5"
Sleep 1300ms

# ── Registry: login modal ───────────────────────────────────────────
Type "6"
Sleep 1200ms
Type "L"
Sleep 1500ms
Escape
Sleep 700ms

# ── System + sub-screens ────────────────────────────────────────────
Type "7"
Sleep 1300ms

Type ":df"
Sleep 300ms
Enter
Sleep 1300ms

Type ":dns"
Sleep 300ms
Enter
Sleep 1300ms

Type ":property"
Sleep 300ms
Enter
Sleep 1300ms

Type ":kernel"
Sleep 300ms
Enter
Sleep 1300ms

# ── k9s parity screens ──────────────────────────────────────────────
Type "8"
Sleep 1500ms
Type "9"
Sleep 1500ms
Type "0"
Sleep 1500ms

Type ":errors"
Sleep 300ms
Enter
Sleep 1500ms

Type ":jobs"
Sleep 300ms
Enter
Sleep 1500ms

# ── New v0.1.x palette commands ─────────────────────────────────────
# Build modal (action)
Type ":build"
Sleep 300ms
Enter
Sleep 1800ms
Escape
Sleep 700ms

# ACR login — show the command being typed (handler will fail gracefully
# in demo because az isn't authenticated; the InfoModal surfaces the
# clear "common fixes" body).
Type ":acr-login myregistry"
Sleep 800ms
Enter
Sleep 2500ms
# Dismiss the info modal
Space
Sleep 700ms

# Docker shim installer — opens the info modal with detected paths.
Type ":install-docker-shim /tmp/c9s-demo-docker"
Sleep 800ms
Enter
Sleep 2500ms
Space
Sleep 800ms

# Cleanup the file we just wrote
Type ":uninstall-docker-shim /tmp/c9s-demo-docker"
Sleep 800ms
Enter
Sleep 1800ms
Space
Sleep 700ms

# ── Back to containers, quit ────────────────────────────────────────
Type "1"
Sleep 1200ms
Type ":q"
Sleep 300ms
Enter
Sleep 800ms
