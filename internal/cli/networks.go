package cli

import (
	"context"
	"encoding/json"
	"strings"
)

// Network represents a parsed network from the CLI.
type Network struct {
	Name       string
	Driver     string
	Subnet     string
	Containers []string
}

// rawNetwork mirrors the JSON shape from `container network ls --format json`.
// Fields are tolerant; missing fields stay zero-valued.
type rawNetwork struct {
	Name       string   `json:"name"`
	Driver     string   `json:"driver"`
	Subnet     string   `json:"subnet"`
	Containers []string `json:"containers"`
}

// ListNetworks implements Client.
func (c *DefaultClient) ListNetworks(ctx context.Context) ([]Network, error) {
	raw, err := runRaw(ctx, c, "cli.list-networks", "network", "ls", "--format", "json")
	if err != nil {
		return nil, err
	}
	return parseNetworks(raw)
}

// CreateNetwork implements Client.
func (c *DefaultClient) CreateNetwork(ctx context.Context, name string) error {
	return runVoid(ctx, c, "cli.create-network", "network/"+name, "network", "create", name)
}

// DeleteNetwork implements Client.
func (c *DefaultClient) DeleteNetwork(ctx context.Context, name string) error {
	return runVoid(ctx, c, "cli.delete-network", "network/"+name, "network", "rm", name)
}

// PruneNetworks implements Client.
func (c *DefaultClient) PruneNetworks(ctx context.Context) (int, error) {
	out, err := runRaw(ctx, c, "cli.prune-networks", "network", "prune")
	if err != nil {
		return 0, err
	}
	return parsePruneCount(out), nil
}

// parseNetworks decodes the JSON output of `container network ls --format json`.
// TODO(plan-4): refine once Apple's network CLI shape is observed in the wild.
func parseNetworks(raw []byte) ([]Network, error) {
	trim := strings.TrimSpace(string(raw))
	if trim == "" || trim == "null" {
		return []Network{}, nil
	}

	var rawList []rawNetwork
	if err := json.Unmarshal(raw, &rawList); err != nil {
		return nil, Wrap("cli.parse-networks", "", err, "failed to decode network list JSON")
	}
	result := make([]Network, 0, len(rawList))
	for _, rn := range rawList {
		result = append(result, Network(rn))
	}
	return result, nil
}
