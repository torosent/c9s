package cli

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// Image represents a parsed image from the CLI.
type Image struct {
	ID         string
	ShortID    string
	Repository string
	Tag        string
	Reference  string
	Created    time.Time
	SizeBytes  int64
}

// rawImage mirrors the JSON shape from `container image ls --format json`.
// Apple's container CLI emits {reference, fullSize, descriptor:{digest, size, mediaType}}.
type rawImage struct {
	Reference  string `json:"reference"`
	FullSize   string `json:"fullSize"`
	Descriptor struct {
		Digest    string `json:"digest"`
		Size      int64  `json:"size"`
		MediaType string `json:"mediaType"`
	} `json:"descriptor"`
}

// ListImages implements Client.
func (c *DefaultClient) ListImages(ctx context.Context) ([]Image, error) {
	raw, err := runRaw(ctx, c, "cli.list-images", "image", "ls", "--format", "json")
	if err != nil {
		return nil, err
	}

	return parseImages(raw)
}

// InspectImage implements Client.
func (c *DefaultClient) InspectImage(ctx context.Context, id string) ([]byte, error) {
	return runRaw(ctx, c, "cli.inspect-image", "image", "inspect", id)
}

// TagImage implements Client.
func (c *DefaultClient) TagImage(ctx context.Context, src, dst string) error {
	return runVoid(ctx, c, "cli.tag-image", "image/"+src, "image", "tag", src, dst)
}

// DeleteImage implements Client.
func (c *DefaultClient) DeleteImage(ctx context.Context, id string) error {
	return runVoid(ctx, c, "cli.delete-image", "image/"+id, "image", "rm", id)
}

// PruneImages implements Client.
func (c *DefaultClient) PruneImages(ctx context.Context, all bool) (int, error) {
	args := []string{"image", "prune"}
	if all {
		args = append(args, "-a")
	}

	out, err := runRaw(ctx, c, "cli.prune-images", args...)
	if err != nil {
		return 0, err
	}

	return parsePruneCount(out), nil
}

// LoadImage implements Client.
func (c *DefaultClient) LoadImage(ctx context.Context, tarPath string) error {
	return runVoid(ctx, c, "cli.load-image", "image", "image", "load", "-i", tarPath)
}

// SaveImage implements Client.
func (c *DefaultClient) SaveImage(ctx context.Context, ref, tarPath string) error {
	return runVoid(ctx, c, "cli.save-image", "image/"+ref, "image", "save", ref, "-o", tarPath)
}

// parseImages decodes the JSON output of `container image ls --format json`
// and converts it to our Image model.
func parseImages(raw []byte) ([]Image, error) {
	var rawList []rawImage
	if err := json.Unmarshal(raw, &rawList); err != nil {
		return nil, Wrap("cli.parse-images", "", err, "failed to decode image list JSON")
	}

	result := make([]Image, 0, len(rawList))
	for _, ri := range rawList {
		result = append(result, projectImage(ri))
	}

	return result, nil
}

// projectImage converts a rawImage to our Image model.
func projectImage(ri rawImage) Image {
	// Apple's CLI emits a full reference like "docker.io/library/nginx:alpine".
	// Split into repository (everything before the last ':') and tag.
	repo, tag := ri.Reference, ""
	if idx := strings.LastIndex(ri.Reference, ":"); idx >= 0 {
		repo = ri.Reference[:idx]
		tag = ri.Reference[idx+1:]
	}

	// ID = the digest, short-stripped.
	id := ri.Descriptor.Digest
	shortID := strings.TrimPrefix(id, "sha256:")
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}

	// Size: prefer parsed bytes from FullSize ("26 MB"), fall back to the
	// descriptor size (which is just the manifest size, not useful for users).
	size := parseHumanSize(ri.FullSize)

	return Image{
		ID:         id,
		ShortID:    shortID,
		Repository: repo,
		Tag:        tag,
		Reference:  ri.Reference,
		Created:    time.Time{}, // CLI doesn't emit a creation time
		SizeBytes:  size,
	}
}

// parseHumanSize converts "26 MB" / "1.4 GB" / "159.3 MB" into bytes.
// Returns 0 on parse failure.
func parseHumanSize(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	parts := strings.Fields(s)
	if len(parts) < 2 {
		return 0
	}
	val, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}
	switch strings.ToUpper(parts[1]) {
	case "B":
		return int64(val)
	case "KB":
		return int64(val * 1024)
	case "MB":
		return int64(val * 1024 * 1024)
	case "GB":
		return int64(val * 1024 * 1024 * 1024)
	case "TB":
		return int64(val * 1024 * 1024 * 1024 * 1024)
	}
	return 0
}
