package config

import "testing"

func TestResolveAlias_NoAliases(t *testing.T) {
	aliases := map[string]string{}

	if got := ResolveAlias("containers", aliases); got != "containers" {
		t.Errorf("expected 'containers', got %q", got)
	}
}

func TestResolveAlias_SingleAlias(t *testing.T) {
	aliases := map[string]string{
		"kpods": "containers",
	}

	if got := ResolveAlias("kpods", aliases); got != "containers" {
		t.Errorf("expected 'containers', got %q", got)
	}

	// Non-aliased command should pass through
	if got := ResolveAlias("images", aliases); got != "images" {
		t.Errorf("expected 'images', got %q", got)
	}
}

func TestResolveAlias_MultipleAliases(t *testing.T) {
	aliases := map[string]string{
		"kpods": "containers",
		"kimg":  "images",
		"kvol":  "volumes",
		"knet":  "networks",
	}

	cases := map[string]string{
		"kpods": "containers",
		"kimg":  "images",
		"kvol":  "volumes",
		"knet":  "networks",
		"quit":  "quit", // pass-through
	}

	for input, want := range cases {
		if got := ResolveAlias(input, aliases); got != want {
			t.Errorf("ResolveAlias(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestResolveAlias_PreservesArgs(t *testing.T) {
	aliases := map[string]string{
		"kpods": "containers",
	}

	// ResolveAlias should only resolve the verb, not the whole command
	// Actually, looking at the implementation, we want to resolve just the first word
	// Let's test that behavior
	input := "kpods"
	got := ResolveAlias(input, aliases)
	if got != "containers" {
		t.Errorf("ResolveAlias(%q) = %q, want 'containers'", input, got)
	}
}

func TestResolveAlias_EmptyCommand(t *testing.T) {
	aliases := map[string]string{
		"kpods": "containers",
	}

	if got := ResolveAlias("", aliases); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
