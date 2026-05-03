package cli

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
)

// SystemService represents a runtime service managed by `container system`.
type SystemService struct {
	Name      string
	State     string
	PID       int
	UptimeSec int64
}

// DFSection represents one section of the `container system df` summary.
type DFSection struct {
	Count        int
	Active       int
	SizeBytes    int64
	ReclaimBytes int64
}

// SystemDF aggregates the three sections returned by `container system df`.
type SystemDF struct {
	Images     DFSection
	Containers DFSection
	Volumes    DFSection
}

// DNSDomain represents a local DNS domain entry.
type DNSDomain struct {
	Name    string
	Default bool
}

// SystemProperty represents a key/value system property.
type SystemProperty struct {
	Key      string
	Value    string
	ReadOnly bool
}

type rawSystemService struct {
	Name      string `json:"name"`
	State     string `json:"state"`
	PID       int    `json:"pid"`
	UptimeSec int64  `json:"uptimeSec"`
}

type rawSystemDF struct {
	Images     rawDFSection `json:"images"`
	Containers rawDFSection `json:"containers"`
	Volumes    rawDFSection `json:"volumes"`
}

type rawDFSection struct {
	Count        int   `json:"count"`
	Active       int   `json:"active"`
	SizeBytes    int64 `json:"size"`
	ReclaimBytes int64 `json:"reclaimable"`
}

type rawDNSDomain struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

type rawSystemProperty struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	ReadOnly bool   `json:"readOnly"`
}

// ListSystemServices implements Client.
func (c *DefaultClient) ListSystemServices(ctx context.Context) ([]SystemService, error) {
	raw, err := runRaw(ctx, c, "cli.list-services", "system", "services", "list", "--format", "json")
	if err != nil {
		return nil, err
	}
	return parseSystemServices(raw)
}

// SystemStartAll implements Client.
func (c *DefaultClient) SystemStartAll(ctx context.Context) error {
	return runVoid(ctx, c, "cli.system-start", "system", "system", "start")
}

// SystemStopAll implements Client.
func (c *DefaultClient) SystemStopAll(ctx context.Context) error {
	return runVoid(ctx, c, "cli.system-stop", "system", "system", "stop")
}

// SystemDF implements Client.
func (c *DefaultClient) SystemDF(ctx context.Context) (SystemDF, error) {
	raw, err := runRaw(ctx, c, "cli.system-df", "system", "df", "--format", "json")
	if err != nil {
		return SystemDF{}, err
	}
	return parseSystemDF(raw)
}

// ListDNSDomains implements Client.
func (c *DefaultClient) ListDNSDomains(ctx context.Context) ([]DNSDomain, error) {
	raw, err := runRaw(ctx, c, "cli.list-dns", "system", "dns", "list", "--format", "json")
	if err != nil {
		return nil, err
	}
	return parseDNSDomains(raw)
}

// CreateDNSDomain implements Client.
func (c *DefaultClient) CreateDNSDomain(ctx context.Context, name string) error {
	return runVoid(ctx, c, "cli.create-dns", "dns/"+name, "system", "dns", "add", name)
}

// DeleteDNSDomain implements Client.
func (c *DefaultClient) DeleteDNSDomain(ctx context.Context, name string) error {
	return runVoid(ctx, c, "cli.delete-dns", "dns/"+name, "system", "dns", "remove", name)
}

// SetDefaultDNSDomain implements Client.
func (c *DefaultClient) SetDefaultDNSDomain(ctx context.Context, name string) error {
	return runVoid(ctx, c, "cli.default-dns", "dns/"+name, "system", "dns", "default", name)
}

// ListSystemProperties implements Client.
func (c *DefaultClient) ListSystemProperties(ctx context.Context) ([]SystemProperty, error) {
	raw, err := runRaw(ctx, c, "cli.list-properties", "system", "property", "list", "--format", "json")
	if err != nil {
		return nil, err
	}
	return parseSystemProperties(raw)
}

// SetSystemProperty implements Client.
func (c *DefaultClient) SetSystemProperty(ctx context.Context, key, value string) error {
	return runVoid(ctx, c, "cli.set-property", "property/"+key, "system", "property", "set", key, value)
}

// ResetSystemProperty implements Client.
func (c *DefaultClient) ResetSystemProperty(ctx context.Context, key string) error {
	return runVoid(ctx, c, "cli.reset-property", "property/"+key, "system", "property", "reset", key)
}

// StreamSystemLogs streams `container system logs` output.
func (c *DefaultClient) StreamSystemLogs(ctx context.Context, follow bool) (Stream, error) {
	args := []string{"system", "logs"}
	if follow {
		args = append(args, "--follow")
	}
	//nolint:gosec // c.bin is hardcoded internally
	cmd := exec.CommandContext(ctx, c.bin, args...)
	return runStream(ctx, cmd, parseLogLine)
}

// parseSystemServices decodes the JSON output of `container system services list --format json`.
// TODO(plan-4): refine once Apple's services CLI shape is observed.
func parseSystemServices(raw []byte) ([]SystemService, error) {
	trim := strings.TrimSpace(string(raw))
	if trim == "" || trim == "null" {
		return []SystemService{}, nil
	}
	var rawList []rawSystemService
	if err := json.Unmarshal(raw, &rawList); err != nil {
		return nil, Wrap("cli.parse-services", "", err, "failed to decode services JSON")
	}
	result := make([]SystemService, 0, len(rawList))
	for _, rs := range rawList {
		result = append(result, SystemService(rs))
	}
	return result, nil
}

// parseSystemDF decodes the JSON output of `container system df --format json`.
// TODO(plan-4): refine once Apple's df CLI shape is observed.
func parseSystemDF(raw []byte) (SystemDF, error) {
	trim := strings.TrimSpace(string(raw))
	if trim == "" || trim == "null" {
		return SystemDF{}, nil
	}
	var rs rawSystemDF
	if err := json.Unmarshal(raw, &rs); err != nil {
		return SystemDF{}, Wrap("cli.parse-df", "", err, "failed to decode df JSON")
	}
	return SystemDF{
		Images:     dfFromRaw(rs.Images),
		Containers: dfFromRaw(rs.Containers),
		Volumes:    dfFromRaw(rs.Volumes),
	}, nil
}

func dfFromRaw(r rawDFSection) DFSection {
	return DFSection(r)
}

// parseDNSDomains decodes the JSON output of `container system dns list --format json`.
func parseDNSDomains(raw []byte) ([]DNSDomain, error) {
	trim := strings.TrimSpace(string(raw))
	if trim == "" || trim == "null" {
		return []DNSDomain{}, nil
	}
	var rawList []rawDNSDomain
	if err := json.Unmarshal(raw, &rawList); err != nil {
		return nil, Wrap("cli.parse-dns", "", err, "failed to decode DNS list JSON")
	}
	result := make([]DNSDomain, 0, len(rawList))
	for _, rd := range rawList {
		result = append(result, DNSDomain(rd))
	}
	return result, nil
}

// parseSystemProperties decodes the JSON output of `container system property list --format json`.
func parseSystemProperties(raw []byte) ([]SystemProperty, error) {
	trim := strings.TrimSpace(string(raw))
	if trim == "" || trim == "null" {
		return []SystemProperty{}, nil
	}
	var rawList []rawSystemProperty
	if err := json.Unmarshal(raw, &rawList); err != nil {
		return nil, Wrap("cli.parse-properties", "", err, "failed to decode properties JSON")
	}
	result := make([]SystemProperty, 0, len(rawList))
	for _, rp := range rawList {
		result = append(result, SystemProperty(rp))
	}
	return result, nil
}
