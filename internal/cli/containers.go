package cli

import (
	"encoding/json"
	"strings"
	"time"
)

// Container represents a parsed container from the CLI.
type Container struct {
	ID        string
	ShortID   string
	Image     string
	Status    string
	StartedAt time.Time
	Uptime    time.Duration
	OS        string
	Arch      string
	Networks  []NetworkAttachment
	Ports     []PortMapping
	CPU       int
	MemBytes  int64
}

// NetworkAttachment represents a network connection.
type NetworkAttachment struct {
	Network  string
	Hostname string
	IP       string
}

// PortMapping represents a published port.
type PortMapping struct {
	Proto         string
	HostPort      int
	ContainerPort int
}

// parseContainers decodes the JSON output of `container ls --format json`
// and converts it to our Container model.
func parseContainers(raw []byte, now time.Time) ([]Container, error) {
	var rawList []rawContainer
	if err := json.Unmarshal(raw, &rawList); err != nil {
		return nil, Wrap("cli.parse-containers", "", err, "failed to decode container list JSON")
	}

	result := make([]Container, 0, len(rawList))
	for _, rc := range rawList {
		result = append(result, projectContainer(rc, now))
	}

	return result, nil
}

// rawContainer mirrors the JSON shape from `container ls --all --format json`.
type rawContainer struct {
	Status        string  `json:"status"`
	StartedDate   float64 `json:"startedDate"` // seconds since 2001-01-01 UTC
	Configuration struct {
		ID    string `json:"id"`
		Image struct {
			Reference string `json:"reference"`
		} `json:"image"`
		Platform struct {
			OS   string `json:"os"`
			Arch string `json:"architecture"`
		} `json:"platform"`
		Resources struct {
			CPUs     int   `json:"cpus"`
			MemBytes int64 `json:"memoryInBytes"`
		} `json:"resources"`
		PublishedPorts []struct {
			Proto         string `json:"proto"`
			HostPort      int    `json:"hostPort"`
			ContainerPort int    `json:"containerPort"`
		} `json:"publishedPorts"`
		Networks []struct {
			Network string `json:"network"`
			Options struct {
				Hostname string `json:"hostname"`
			} `json:"options"`
		} `json:"networks"`
	} `json:"configuration"`
	Networks []struct {
		Network     string `json:"network"`
		Hostname    string `json:"hostname"`
		IPv4Address string `json:"ipv4Address"`
	} `json:"networks"`
}

// projectContainer converts a rawContainer to our Container model.
func projectContainer(rc rawContainer, now time.Time) Container {
	// Apple's epoch: 2001-01-01 00:00:00 UTC
	appleEpoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	startedAt := appleEpoch.Add(time.Duration(rc.StartedDate * float64(time.Second)))

	// Calculate uptime only for running containers
	var uptime time.Duration
	if strings.ToLower(rc.Status) == "running" {
		uptime = now.Sub(startedAt)
	}

	// Extract networks from the top-level networks (which has IPs)
	var networks []NetworkAttachment
	for _, net := range rc.Networks {
		networks = append(networks, NetworkAttachment{
			Network:  net.Network,
			Hostname: net.Hostname,
			IP:       net.IPv4Address,
		})
	}
	if len(networks) == 0 {
		// Fall back to configuration.networks (no IP info)
		for _, net := range rc.Configuration.Networks {
			networks = append(networks, NetworkAttachment{
				Network:  net.Network,
				Hostname: net.Options.Hostname,
				IP:       "",
			})
		}
	}

	// Extract ports
	var ports []PortMapping
	for _, port := range rc.Configuration.PublishedPorts {
		ports = append(ports, PortMapping{
			Proto:         port.Proto,
			HostPort:      port.HostPort,
			ContainerPort: port.ContainerPort,
		})
	}

	id := rc.Configuration.ID
	// Generate short ID: remove dashes and take first 12 chars
	shortID := strings.ReplaceAll(id, "-", "")
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}

	return Container{
		ID:        id,
		ShortID:   shortID,
		Image:     rc.Configuration.Image.Reference,
		Status:    rc.Status,
		StartedAt: startedAt,
		Uptime:    uptime,
		OS:        rc.Configuration.Platform.OS,
		Arch:      rc.Configuration.Platform.Arch,
		Networks:  networks,
		Ports:     ports,
		CPU:       rc.Configuration.Resources.CPUs,
		MemBytes:  rc.Configuration.Resources.MemBytes,
	}
}

// timeNow returns the current time; extracted for testing.
var timeNow = time.Now

// parsePruneCount extracts the count from "N containers removed" output.
func parsePruneCount(out []byte) int {
	// Best-effort regex match
	s := string(out)
	// Look for pattern like "3 containers removed"
	var count int
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			n := 0
			for i < len(s) && s[i] >= '0' && s[i] <= '9' {
				n = n*10 + int(s[i]-'0')
				i++
			}
			count = n
			break
		}
	}
	return count
}
