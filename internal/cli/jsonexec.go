package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"strings"
)

// runJSON runs the container CLI with the given args, captures stdout and stderr,
// and decodes the stdout as JSON into type T.
// On non-zero exit or decode failure, returns wrapped error with operation context.
func runJSON[T any](ctx context.Context, c *DefaultClient, op string, args ...string) (T, error) {
	var zero T

	//nolint:gosec // c.bin is controlled by c9s, not user input
	cmd := exec.CommandContext(ctx, c.bin, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrTrim := strings.TrimSpace(stderr.String())
		return zero, Wrap(op, "", err, stderrTrim)
	}

	var result T
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return zero, Wrap(op, "", err, "could not parse `container ... --format json` output")
	}

	return result, nil
}
