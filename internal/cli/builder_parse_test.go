package cli

import "testing"

func TestParseBuilderStatus_PlainTextNotRunning(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"builder is not running", "stopped"},
		{"BUILDER IS NOT RUNNING", "stopped"},
		{"  builder is not running  ", "stopped"},
		{"builder is running", "running"},
		{"builder is stopped", "stopped"},
	}
	for _, tc := range cases {
		got, err := parseBuilderStatus([]byte(tc.in))
		if err != nil {
			t.Errorf("parseBuilderStatus(%q) errored: %v", tc.in, err)
			continue
		}
		if got.State != tc.want {
			t.Errorf("parseBuilderStatus(%q).State = %q, want %q", tc.in, got.State, tc.want)
		}
	}
}

func TestParseBuilderStatus_Empty(t *testing.T) {
	got, err := parseBuilderStatus([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.State != "unknown" {
		t.Errorf("empty input should yield 'unknown', got %q", got.State)
	}
}

func TestParseBuilderStatus_JSON(t *testing.T) {
	in := `{"state":"running","cpus":4,"memoryBytes":2147483648,"uptimeSec":120}`
	got, err := parseBuilderStatus([]byte(in))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.State != "running" {
		t.Errorf("State=%q, want running", got.State)
	}
	if got.CPUs != 4 {
		t.Errorf("CPUs=%d, want 4", got.CPUs)
	}
}

func TestParseBuilderStatus_BadJSON(t *testing.T) {
	in := `{"state":"running"`
	if _, err := parseBuilderStatus([]byte(in)); err == nil {
		t.Error("malformed JSON should produce an error")
	}
}
