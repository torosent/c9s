package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"strings"
)

// RegistryEntry represents a configured registry login.
type RegistryEntry struct {
	Host    string
	User    string
	Default bool
}

// rawRegistryEntry mirrors the JSON shape from `container registry list --format json`.
type rawRegistryEntry struct {
	Host    string `json:"host"`
	User    string `json:"user"`
	Default bool   `json:"default"`
}

// ListRegistries implements Client.
func (c *DefaultClient) ListRegistries(ctx context.Context) ([]RegistryEntry, error) {
	raw, err := runRaw(ctx, c, "cli.list-registries", "registry", "list", "--format", "json")
	if err != nil {
		return nil, err
	}
	return parseRegistries(raw)
}

// RegistryLogin implements Client. The password is sent via stdin so it
// never appears on the process argv list.
func (c *DefaultClient) RegistryLogin(ctx context.Context, host, user, pass string) error {
	args := []string{"registry", "login", host, "--username", user, "--password-stdin"}
	//nolint:gosec // c.bin is hardcoded internally; user inputs are isolated to args/stdin
	cmd := exec.CommandContext(ctx, c.bin, args...)
	cmd.Stdin = strings.NewReader(pass + "\n")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return Wrap("cli.registry-login", "registry/"+host, err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// RegistryLogout implements Client.
func (c *DefaultClient) RegistryLogout(ctx context.Context, host string) error {
	return runVoid(ctx, c, "cli.registry-logout", "registry/"+host, "registry", "logout", host)
}

// RegistrySetDefault implements Client.
func (c *DefaultClient) RegistrySetDefault(ctx context.Context, host string) error {
	return runVoid(ctx, c, "cli.registry-default", "registry/"+host, "registry", "default", "set", host)
}

// parseRegistries decodes the JSON output of `container registry list --format json`.
// TODO(plan-4): refine once Apple's registry CLI shape is observed.
func parseRegistries(raw []byte) ([]RegistryEntry, error) {
	trim := strings.TrimSpace(string(raw))
	if trim == "" || trim == "null" {
		return []RegistryEntry{}, nil
	}
	var rawList []rawRegistryEntry
	if err := json.Unmarshal(raw, &rawList); err != nil {
		return nil, Wrap("cli.parse-registries", "", err, "failed to decode registry list JSON")
	}
	result := make([]RegistryEntry, 0, len(rawList))
	for _, rr := range rawList {
		result = append(result, RegistryEntry(rr))
	}
	return result, nil
}
