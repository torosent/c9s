package cli

import (
	"context"
	"regexp"
	"strconv"
)

// Capabilities describes optional features of the `container` runtime.
// In v0.1.0 we only parse the version; per-feature flags will land as
// later versions of the runtime introduce them.
type Capabilities struct {
	// Version holds the semver substring (e.g. "0.12.1") when `container --version`
	// could be parsed, or the raw stdout otherwise. Downstream UI can render
	// this directly without further trimming.
	Version string
	Major   int
	Minor   int
	Patch   int

	Pause bool // reserved for future use
}

var versionRe = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)

// ProbeCapabilities reads `container --version` via the supplied Client
// and parses the result.
func ProbeCapabilities(ctx context.Context, c Client) (Capabilities, error) {
	raw, err := c.Version(ctx)
	if err != nil {
		return Capabilities{}, Wrap("cli.probe-capabilities", "", err, "is `container` installed and on PATH?")
	}
	caps := Capabilities{Version: raw}
	if m := versionRe.FindStringSubmatch(raw); len(m) == 4 {
		caps.Major, _ = strconv.Atoi(m[1])
		caps.Minor, _ = strconv.Atoi(m[2])
		caps.Patch, _ = strconv.Atoi(m[3])
		caps.Version = m[0] // overwrite raw with just the semver
	}
	return caps, nil
}
