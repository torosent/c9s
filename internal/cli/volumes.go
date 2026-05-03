package cli

import (
	"context"
	"encoding/json"
)

// Volume represents a parsed volume from the CLI.
type Volume struct {
	Name       string
	Driver     string
	Mountpoint string
	SizeBytes  int64
	UsedBy     []string
}

// rawVolume mirrors the JSON shape from `container volume ls --format json`.
// The exact shape is uncertain across CLI versions; we tolerate missing fields.
type rawVolume struct {
	Name       string   `json:"name"`
	Driver     string   `json:"driver"`
	Mountpoint string   `json:"mountpoint"`
	Size       int64    `json:"size"`
	UsedBy     []string `json:"usedBy"`
}

// ListVolumes implements Client.
func (c *DefaultClient) ListVolumes(ctx context.Context) ([]Volume, error) {
	raw, err := runRaw(ctx, c, "cli.list-volumes", "volume", "ls", "--format", "json")
	if err != nil {
		return nil, err
	}
	return parseVolumes(raw)
}

// CreateVolume implements Client.
func (c *DefaultClient) CreateVolume(ctx context.Context, name string) error {
	return runVoid(ctx, c, "cli.create-volume", "volume/"+name, "volume", "create", name)
}

// DeleteVolume implements Client.
func (c *DefaultClient) DeleteVolume(ctx context.Context, name string) error {
	return runVoid(ctx, c, "cli.delete-volume", "volume/"+name, "volume", "rm", name)
}

// PruneVolumes implements Client.
func (c *DefaultClient) PruneVolumes(ctx context.Context) (int, error) {
	out, err := runRaw(ctx, c, "cli.prune-volumes", "volume", "prune")
	if err != nil {
		return 0, err
	}
	return parsePruneCount(out), nil
}

// parseVolumes decodes the JSON output of `container volume ls --format json`.
// TODO(plan-4): refine once Apple's volume CLI shape is observed in the wild.
func parseVolumes(raw []byte) ([]Volume, error) {
	var rawList []rawVolume
	if err := json.Unmarshal(raw, &rawList); err != nil {
		return nil, Wrap("cli.parse-volumes", "", err, "failed to decode volume list JSON")
	}
	result := make([]Volume, 0, len(rawList))
	for _, rv := range rawList {
		result = append(result, Volume{
			Name:       rv.Name,
			Driver:     rv.Driver,
			Mountpoint: rv.Mountpoint,
			SizeBytes:  rv.Size,
			UsedBy:     rv.UsedBy,
		})
	}
	return result, nil
}
