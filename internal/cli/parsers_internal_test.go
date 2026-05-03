package cli

import (
	"errors"
	"testing"
)

func TestParseHumanSize(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"", 0},
		{"500 B", 500},
		// Binary semantics — see images.go docstring + testdata/image-ls.json.
		{"1 KB", 1024},
		{"1 MB", 1024 * 1024},
		{"500 MB", 500 * 1024 * 1024},
		{"1.4 GB", 1503238553},
		{"2 TB", 2 * 1024 * 1024 * 1024 * 1024},
		// Explicit IEC suffixes are also accepted (same multipliers).
		{"1 KiB", 1024},
		{"1 MiB", 1024 * 1024},
		{"1 GiB", 1024 * 1024 * 1024},
		{"garbage", 0},
		{"100 ZB", 0}, // unknown unit
		{"abc MB", 0},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := parseHumanSize(c.in)
			if got != c.want {
				t.Errorf("parseHumanSize(%q) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}

func TestParsePruneCount_NoSha256Confusion(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		// Legacy cases that already worked.
		{"removed-3-containers", "Removed 3 containers", 3},
		{"3-containers-removed", "3 containers removed.", 3},
		{"images-plural", "Removed 12 images", 12},
		{"single-network", "1 network removed", 1},
		{"volumes", "Removed 7 volumes", 7},
		// Regression: the old implementation greedily grabbed the first
		// run of digits, which would mis-parse this as 256. The regex
		// pins to a unit word.
		{"sha256-noise", "Removed sha256:abc123def 5 containers", 5},
		{"no-unit", "ID: 12345", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := parsePruneCount([]byte(c.in)); got != c.want {
				t.Errorf("parsePruneCount(%q) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}

func TestErrUnsupported_IsCheck(t *testing.T) {
	wrapped := Wrap("cli.pause-unsupported", "container/abc", ErrUnsupported, "no pause")
	if !errors.Is(wrapped, ErrUnsupported) {
		t.Fatal("errors.Is(err, ErrUnsupported) should be true for wrapped sentinel")
	}
}
