package cli

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

// Client is the gateway to Apple's `container` CLI. It is the only
// interface the rest of c9s uses to talk to the runtime.
type Client interface {
	// Capabilities probes the runtime for feature flags.
	Capabilities(ctx context.Context) (Capabilities, error)
	// Version returns the raw stdout of `container --version`.
	Version(ctx context.Context) (string, error)

	// Plan 2 additions: container lifecycle methods
	ListContainers(ctx context.Context, all bool) ([]Container, error)
	InspectContainer(ctx context.Context, id string) ([]byte, error)
	StopContainer(ctx context.Context, id string) error
	KillContainer(ctx context.Context, id string) error
	RestartContainer(ctx context.Context, id string) error
	DeleteContainer(ctx context.Context, id string) error
	PauseContainer(ctx context.Context, id string) error
	UnpauseContainer(ctx context.Context, id string) error
	PruneContainers(ctx context.Context) (int, error)

	// Plan 3 additions: streaming methods
	StreamLogs(ctx context.Context, id string, follow bool) (Stream, error)
	StreamBuild(ctx context.Context, opts BuildOpts) (Stream, error)
	StreamPull(ctx context.Context, ref string) (Stream, error)
	StreamPush(ctx context.Context, ref string) (Stream, error)

	// Plan 4 additions: image methods
	ListImages(ctx context.Context) ([]Image, error)
	InspectImage(ctx context.Context, id string) ([]byte, error)
	TagImage(ctx context.Context, src, dst string) error
	DeleteImage(ctx context.Context, id string) error
	PruneImages(ctx context.Context, all bool) (int, error)
	LoadImage(ctx context.Context, tarPath string) error
	SaveImage(ctx context.Context, ref, tarPath string) error

	// Plan 4 additions: volume methods
	ListVolumes(ctx context.Context) ([]Volume, error)
	CreateVolume(ctx context.Context, name string) error
	DeleteVolume(ctx context.Context, name string) error
	PruneVolumes(ctx context.Context) (int, error)

	// Plan 4 additions: network methods
	ListNetworks(ctx context.Context) ([]Network, error)
	CreateNetwork(ctx context.Context, name string) error
	DeleteNetwork(ctx context.Context, name string) error
	PruneNetworks(ctx context.Context) (int, error)

	// Plan 4 additions: builder methods
	BuilderStatus(ctx context.Context) (BuilderStatus, error)
	BuilderStart(ctx context.Context) error
	BuilderStop(ctx context.Context) error
	BuilderDelete(ctx context.Context) error

	// Plan 4 additions: registry methods
	ListRegistries(ctx context.Context) ([]RegistryEntry, error)
	RegistryLogin(ctx context.Context, host, user, pass string) error
	RegistryLogout(ctx context.Context, host string) error
	RegistrySetDefault(ctx context.Context, host string) error

	// Plan 4 additions: system methods
	ListSystemServices(ctx context.Context) ([]SystemService, error)
	SystemStartAll(ctx context.Context) error
	SystemStopAll(ctx context.Context) error
	SystemDF(ctx context.Context) (SystemDF, error)
	ListDNSDomains(ctx context.Context) ([]DNSDomain, error)
	CreateDNSDomain(ctx context.Context, name string) error
	DeleteDNSDomain(ctx context.Context, name string) error
	SetDefaultDNSDomain(ctx context.Context, name string) error
	ListSystemProperties(ctx context.Context) ([]SystemProperty, error)
	SetSystemProperty(ctx context.Context, key, value string) error
	ResetSystemProperty(ctx context.Context, key string) error
	StreamSystemLogs(ctx context.Context, follow bool) (Stream, error)

	// Plan 4 additions: run
	RunContainer(ctx context.Context, opts RunOpts) (Stream, error)
}

// DefaultClient is the production implementation that shells out to the
// `container` binary on PATH.
type DefaultClient struct {
	bin string
}

// Option configures a DefaultClient.
type Option func(*DefaultClient)

// WithBinary overrides the path to the `container` binary.
// Tests typically pass `echo` or a fixture script.
func WithBinary(path string) Option {
	return func(c *DefaultClient) { c.bin = path }
}

// NewDefaultClient returns a DefaultClient whose binary defaults to "container".
func NewDefaultClient(opts ...Option) *DefaultClient {
	c := &DefaultClient{bin: "container"}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Version implements Client.
func (c *DefaultClient) Version(ctx context.Context) (string, error) {
	//nolint:gosec // c.bin is hardcoded internally, not user-controlled
	cmd := exec.CommandContext(ctx, c.bin, "--version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", Wrap("cli.version", "", err, msg)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Capabilities implements Client.
func (c *DefaultClient) Capabilities(ctx context.Context) (Capabilities, error) {
	return ProbeCapabilities(ctx, c)
}

// runRaw runs a command and returns stdout bytes on success, or wrapped error on failure.
func runRaw(ctx context.Context, c *DefaultClient, op string, args ...string) ([]byte, error) {
	//nolint:gosec // c.bin is controlled by c9s, not user input
	cmd := exec.CommandContext(ctx, c.bin, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrTrim := strings.TrimSpace(stderr.String())
		return nil, Wrap(op, "", err, stderrTrim)
	}

	return stdout.Bytes(), nil
}

// runVoid runs a command and returns nil on success or wrapped error on failure.
// Stdout is captured and surfaced via the wrapped error's Hint when stderr is
// empty, since some `container` subcommands emit failure context on stdout.
func runVoid(ctx context.Context, c *DefaultClient, op, resource string, args ...string) error {
	//nolint:gosec // c.bin is controlled by c9s, not user input
	cmd := exec.CommandContext(ctx, c.bin, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		hint := strings.TrimSpace(stderr.String())
		if hint == "" {
			hint = strings.TrimSpace(stdout.String())
		}
		return Wrap(op, resource, err, hint)
	}

	return nil
}

// ListContainers implements Client.
func (c *DefaultClient) ListContainers(ctx context.Context, all bool) ([]Container, error) {
	args := []string{"ls", "--format", "json"}
	if all {
		args = append(args, "--all")
	}

	raw, err := runRaw(ctx, c, "cli.list-containers", args...)
	if err != nil {
		return nil, err
	}

	return parseContainers(raw, timeNow())
}

// InspectContainer implements Client.
func (c *DefaultClient) InspectContainer(ctx context.Context, id string) ([]byte, error) {
	return runRaw(ctx, c, "cli.inspect-container", "inspect", id)
}

// StopContainer implements Client.
func (c *DefaultClient) StopContainer(ctx context.Context, id string) error {
	return runVoid(ctx, c, "cli.stop-container", "container/"+id, "stop", id)
}

// KillContainer implements Client.
func (c *DefaultClient) KillContainer(ctx context.Context, id string) error {
	return runVoid(ctx, c, "cli.kill-container", "container/"+id, "kill", id)
}

// RestartContainer implements Client.
func (c *DefaultClient) RestartContainer(ctx context.Context, id string) error {
	return runVoid(ctx, c, "cli.restart-container", "container/"+id, "restart", id)
}

// DeleteContainer implements Client. Uses --force so that running containers
// can be deleted in one step (k9s-style behavior).
func (c *DefaultClient) DeleteContainer(ctx context.Context, id string) error {
	return runVoid(ctx, c, "cli.delete-container", "container/"+id, "delete", "--force", id)
}

// PauseContainer implements Client.
func (c *DefaultClient) PauseContainer(ctx context.Context, id string) error {
	return Wrap("cli.pause-unsupported", "container/"+id, ErrUnsupported, "this container CLI version does not support pause")
}

// UnpauseContainer implements Client.
func (c *DefaultClient) UnpauseContainer(ctx context.Context, id string) error {
	return Wrap("cli.pause-unsupported", "container/"+id, ErrUnsupported, "this container CLI version does not support unpause")
}

// PruneContainers implements Client.
func (c *DefaultClient) PruneContainers(ctx context.Context) (int, error) {
	out, err := runRaw(ctx, c, "cli.prune-containers", "prune")
	if err != nil {
		return 0, err
	}

	return parsePruneCount(out), nil
}
