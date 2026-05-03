# c9s — comprehensive demo against the demo dataset.
# Showcases every screen with the latest UI (light skin, hand-rolled
# pickers, all bg painting fixes).
Output docs/assets/c9s.gif

Set Theme "Builtin Solarized Light"
Set Width 1400
Set Height 800
Set FontSize 14
Set Padding 20
Set TypingSpeed 40ms

Type "c9s --demo-data"
Sleep 400ms
Enter
Sleep 1500ms

# Splash screen
Sleep 2000ms
Space
Sleep 1500ms

# Containers screen (default landing) — light skin pre-loaded, table
# rendered inside the rounded BorderedBox with k9s-style selection row.
Sleep 1500ms

# Switch skin via the interactive picker (hand-rolled, no bg leaks).
Type ":skins"
Sleep 400ms
Enter
Sleep 1500ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Up
Sleep 400ms
Escape
Sleep 1000ms

# Apply a specific skin from the palette.
Type ":skin k9s-light"
Sleep 400ms
Enter
Sleep 1500ms

# Number shortcuts: jump between resource screens.
Type "2"
Sleep 1500ms
Type "3"
Sleep 1500ms
Type "4"
Sleep 1500ms
Type "5"
Sleep 1500ms
Type "6"
Sleep 1500ms
Type "7"
Sleep 1500ms
Type "8"
Sleep 1500ms

# Back to containers.
Type "1"
Sleep 1500ms

# Sort picker (hand-rolled list, accent-highlighted selection).
Type "S"
Sleep 1500ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Type "r"
Sleep 600ms
Escape
Sleep 1000ms

# Filter
Type "/"
Sleep 400ms
Type "api"
Sleep 800ms
Escape
Sleep 800ms

# Help modal
Type "?"
Sleep 1500ms
Escape
Sleep 800ms

# Images screen
Type ":images"
Sleep 400ms
Enter
Sleep 1500ms

# Run modal from images (Enter on focused image)
Down
Sleep 400ms
Enter
Sleep 2000ms
Escape
Sleep 800ms

# Build modal
Type ":build"
Sleep 400ms
Enter
Sleep 2000ms
Escape
Sleep 800ms

# Builder card
Type ":builder"
Sleep 400ms
Enter
Sleep 1500ms

# System info screen
Type ":system"
Sleep 400ms
Enter
Sleep 1500ms

# DF (disk usage)
Type ":df"
Sleep 400ms
Enter
Sleep 1500ms

# Pulses dashboard
Type ":pulses"
Sleep 400ms
Enter
Sleep 1500ms

# Pinned bookmarks
Type ":pinned"
Sleep 400ms
Enter
Sleep 1500ms

# Xray relationships
Type ":xray"
Sleep 400ms
Enter
Sleep 1500ms

# Errors log
Type ":errors"
Sleep 400ms
Enter
Sleep 1500ms

# Back to containers, finish.
Type "1"
Sleep 1500ms
Sleep 1000ms
