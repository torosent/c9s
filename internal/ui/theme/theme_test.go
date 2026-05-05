package theme

import "testing"

func TestDefaultDarkHasRequiredKeys(t *testing.T) {
	p := DefaultDark()
	if p.Fg == nil || p.Bg == nil {
		t.Error("Fg/Bg must be set")
	}
	if p.Accent == nil || p.Error == nil {
		t.Error("Accent/Error must be set")
	}
	for _, key := range []string{"running", "exited", "paused", "stopping", "created"} {
		if _, ok := p.State[key]; !ok {
			t.Errorf("State[%q] missing", key)
		}
	}
}
