package version

import (
	"strings"
	"testing"
)

func TestStringDefault(t *testing.T) {
	got := String()
	if !strings.Contains(got, "c9s") {
		t.Fatalf("expected version string to contain 'c9s', got %q", got)
	}
	if !strings.Contains(got, Version) {
		t.Fatalf("expected version string to contain Version=%q, got %q", Version, got)
	}
	if !strings.Contains(got, Commit) {
		t.Fatalf("expected version string to contain Commit=%q, got %q", Commit, got)
	}
}

func TestStringWithLdflagOverrides(t *testing.T) {
	saveV, saveC, saveD := Version, Commit, Date
	Version, Commit, Date = "0.1.0", "abc1234", "2026-05-02"
	defer func() { Version, Commit, Date = saveV, saveC, saveD }()

	got := String()
	want := "c9s 0.1.0 (commit abc1234, built 2026-05-02)"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
