// Package version exposes build-time information injected via ldflags.
package version

import "fmt"

var (
	// Version is the semver release; "dev" for unreleased builds.
	Version = "dev"
	// Commit is the abbreviated git SHA the binary was built from.
	Commit = "none"
	// Date is the build date (RFC3339 or YYYY-MM-DD).
	Date = "unknown"
)

// Short returns the version string only (no commit/date).
func Short() string { return Version }

// String returns a human-readable identification line.
func String() string {
	return fmt.Sprintf("c9s %s (commit %s, built %s)", Version, Commit, Date)
}
