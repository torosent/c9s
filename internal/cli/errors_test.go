package cli

import (
	"errors"
	"strings"
	"testing"
)

func TestWrapAndUnwrap(t *testing.T) {
	cause := errors.New("exit status 1")
	err := Wrap("cli.list-containers", "container/api-server", cause, "is the service running?")

	var c9err *Error
	if !errors.As(err, &c9err) {
		t.Fatalf("expected *Error from Wrap, got %T", err)
	}
	if c9err.Op != "cli.list-containers" {
		t.Errorf("Op = %q", c9err.Op)
	}
	if c9err.Resource != "container/api-server" {
		t.Errorf("Resource = %q", c9err.Resource)
	}
	if c9err.Hint != "is the service running?" {
		t.Errorf("Hint = %q", c9err.Hint)
	}
	if !errors.Is(err, cause) {
		t.Errorf("errors.Is should match wrapped cause")
	}
}

func TestErrorMessage(t *testing.T) {
	err := Wrap("cli.stop", "container/foo", errors.New("no such container"), "")
	msg := err.Error()
	for _, want := range []string{"cli.stop", "container/foo", "no such container"} {
		if !strings.Contains(msg, want) {
			t.Errorf("Error() missing %q: %s", want, msg)
		}
	}
}

func TestErrorMessageExactFormat(t *testing.T) {
	cases := []struct {
		name       string
		op, res    string
		cause      error
		hint, want string
	}{
		{
			name:  "with_hint",
			op:    "cli.stop",
			res:   "container/foo",
			cause: errors.New("no such container"),
			hint:  "check logs",
			want:  "cli.stop on container/foo: no such container (hint: check logs)",
		},
		{
			name:  "without_hint",
			op:    "cli.list-containers",
			res:   "container/api-server",
			cause: errors.New("timeout"),
			hint:  "",
			want:  "cli.list-containers on container/api-server: timeout",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Wrap(tc.op, tc.res, tc.cause, tc.hint)
			if got := err.Error(); got != tc.want {
				t.Errorf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestWrapNilCauseReturnsNil(t *testing.T) {
	if got := Wrap("op", "res", nil, "hint"); got != nil {
		t.Errorf("Wrap(_, _, nil, _) = %v, want nil", got)
	}
}
