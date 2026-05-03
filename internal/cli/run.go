package cli

import (
	"context"
	"os/exec"
	"strings"
)

// RunOpts contains parameters for `container run`.
type RunOpts struct {
	Name        string
	Image       string
	Ports       []string // e.g. ["8080:8080", "443:443"]
	Env         []string // e.g. ["KEY=val"]
	Volumes     []string // e.g. ["src:dst"]
	Interactive bool
	TTY         bool
	Detach      bool
	ExtraArgs   []string
}

// args converts RunOpts to command-line arguments. The image must be the
// last positional arg before any user-provided extras.
func (opts RunOpts) args() []string {
	args := []string{"run"}
	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}
	for _, p := range opts.Ports {
		if p == "" {
			continue
		}
		args = append(args, "-p", p)
	}
	for _, e := range opts.Env {
		if e == "" {
			continue
		}
		args = append(args, "-e", e)
	}
	for _, v := range opts.Volumes {
		if v == "" {
			continue
		}
		args = append(args, "-v", v)
	}
	if opts.Interactive {
		args = append(args, "-i")
	}
	if opts.TTY {
		args = append(args, "-t")
	}
	if opts.Detach {
		args = append(args, "-d")
	}
	args = append(args, opts.ExtraArgs...)
	args = append(args, opts.Image)
	return args
}

// RunContainer streams output from `container run`.
func (c *DefaultClient) RunContainer(ctx context.Context, opts RunOpts) (Stream, error) {
	args := opts.args()
	//nolint:gosec // c.bin is hardcoded internally
	cmd := exec.CommandContext(ctx, c.bin, args...)
	return runStream(ctx, cmd, parseRunLine)
}

// parseRunLine treats output from `container run` as raw lines so the
// progress modal can echo them. Streaming is light here because run
// returns quickly when -d is set, and otherwise emits container output.
func parseRunLine(line string) []StreamEvent {
	trim := strings.TrimSpace(line)
	if trim == "" {
		return []StreamEvent{RawLine{Text: line}}
	}
	return []StreamEvent{RawLine{Text: line}}
}
