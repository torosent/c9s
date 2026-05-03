package cli

import (
	"context"
	"encoding/json"
	"strings"
)

// BuilderStatus represents the current state of the build subsystem.
type BuilderStatus struct {
	State       string
	CPUs        int
	MemoryBytes int64
	UptimeSec   int64
}

// rawBuilderStatus is the JSON shape from `container builder status --format json`.
type rawBuilderStatus struct {
	State       string `json:"state"`
	CPUs        int    `json:"cpus"`
	MemoryBytes int64  `json:"memoryBytes"`
	UptimeSec   int64  `json:"uptimeSec"`
}

// BuilderStatus implements Client.
func (c *DefaultClient) BuilderStatus(ctx context.Context) (BuilderStatus, error) {
	raw, err := runRaw(ctx, c, "cli.builder-status", "builder", "status", "--format", "json")
	if err != nil {
		return BuilderStatus{}, err
	}
	return parseBuilderStatus(raw)
}

// BuilderStart implements Client.
func (c *DefaultClient) BuilderStart(ctx context.Context) error {
	return runVoid(ctx, c, "cli.builder-start", "builder", "builder", "start")
}

// BuilderStop implements Client.
func (c *DefaultClient) BuilderStop(ctx context.Context) error {
	return runVoid(ctx, c, "cli.builder-stop", "builder", "builder", "stop")
}

// BuilderDelete implements Client.
func (c *DefaultClient) BuilderDelete(ctx context.Context) error {
	return runVoid(ctx, c, "cli.builder-delete", "builder", "builder", "delete")
}

// Apple's CLI may emit JSON (when the builder is running) or a plain text
// status line like "builder is not running" (when stopped). Handle both.
func parseBuilderStatus(raw []byte) (BuilderStatus, error) {
	trim := strings.TrimSpace(string(raw))
	if trim == "" || trim == "null" {
		return BuilderStatus{State: "unknown"}, nil
	}

	// Plain-text status line takes precedence over JSON because the CLI
	// emits it even with --format json when the daemon isn't running.
	lower := strings.ToLower(trim)
	if !strings.HasPrefix(trim, "{") && !strings.HasPrefix(trim, "[") {
		switch {
		case strings.Contains(lower, "not running"), strings.Contains(lower, "stopped"):
			return BuilderStatus{State: "stopped"}, nil
		case strings.Contains(lower, "running"):
			return BuilderStatus{State: "running"}, nil
		default:
			return BuilderStatus{State: trim}, nil
		}
	}

	var rs rawBuilderStatus
	if err := json.Unmarshal(raw, &rs); err != nil {
		return BuilderStatus{}, Wrap("cli.parse-builder", "", err, "failed to decode builder status JSON")
	}
	if rs.State == "" {
		rs.State = "running"
	}
	return BuilderStatus(rs), nil
}
