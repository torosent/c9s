package palette

import "testing"

func TestCatalog_HasCanonicalScreens(t *testing.T) {
	cat := Catalog()
	if len(cat) == 0 {
		t.Fatal("Catalog is empty")
	}
	required := []string{"containers", "images", "volumes", "networks", "builder", "registry", "system", "pulses", "xray", "skin", "skins", "quit"}
	have := make(map[string]bool, len(cat))
	for _, c := range cat {
		have[c.Name] = true
	}
	for _, r := range required {
		if !have[r] {
			t.Errorf("Catalog missing required command %q", r)
		}
	}
}

func TestCatalog_AliasesAreUnique(t *testing.T) {
	seen := map[string]string{}
	for _, c := range Catalog() {
		if other, dup := seen[c.Name]; dup {
			t.Errorf("duplicate name %q (also in %q)", c.Name, other)
		}
		seen[c.Name] = c.Name
		for _, a := range c.Aliases {
			if other, dup := seen[a]; dup {
				t.Errorf("alias %q on %q collides with %q", a, c.Name, other)
			}
			seen[a] = c.Name
		}
	}
}

func TestMatch_EmptyPrefixReturnsAll(t *testing.T) {
	cat := Catalog()
	got := Match("", cat)
	if len(got) != len(cat) {
		t.Errorf("empty prefix: expected %d, got %d", len(cat), len(got))
	}
}

func TestMatch_ExactNamesFirst(t *testing.T) {
	cat := []Command{
		{Name: "skin"},
		{Name: "skins"},
	}
	got := Match("skin", cat)
	if len(got) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(got))
	}
	if got[0].Name != "skin" {
		t.Errorf("expected exact match 'skin' first, got %q", got[0].Name)
	}
}

func TestMatch_PrefixCaseInsensitive(t *testing.T) {
	cat := Catalog()
	got := Match("CONT", cat)
	if len(got) == 0 {
		t.Fatal("expected at least one prefix match for 'CONT'")
	}
	found := false
	for _, c := range got {
		if c.Name == "containers" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'containers' in matches for 'CONT'")
	}
}

func TestMatch_AliasMatch(t *testing.T) {
	cat := Catalog()
	got := Match("c", cat)
	// 'c' is the alias for containers; should appear (either by name or alias)
	found := false
	for _, c := range got {
		if c.Name == "containers" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'containers' to match 'c' via its alias")
	}
}

func TestMatch_NoMatch(t *testing.T) {
	got := Match("zzzz", Catalog())
	if len(got) != 0 {
		t.Errorf("expected 0 matches for nonsense prefix, got %d", len(got))
	}
}

func TestHasPrefix(t *testing.T) {
	cases := []struct {
		s, prefix string
		want      bool
	}{
		{"containers", "cont", true},
		{"containers", "CONT", true},
		{"Cont", "cont", true},
		{"foo", "bar", false},
		{"hi", "hello", false},
	}
	for _, tc := range cases {
		got := hasPrefix(tc.s, tc.prefix)
		if got != tc.want {
			t.Errorf("hasPrefix(%q,%q)=%v, want %v", tc.s, tc.prefix, got, tc.want)
		}
	}
}
