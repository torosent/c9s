// Package demodata provides a populated cli.Fake for demo mode.
package demodata

import (
	"fmt"
	"time"

	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
)

// NewFake constructs a cli.Fake populated with realistic demo data for
// 12 containers, 8 images, 4 volumes, 3 networks, a running builder,
// 2 registry hosts, 1 DNS domain, and 5 system services. Streaming
// functions replay scripted sequences.
func NewFake(c clock.Clock) *cli.Fake {
	now := c.Now()

	fake := cli.NewFake()

	// --- 12 containers: 8 running, 3 stopped, 1 paused ---
	fake.ListContainersResp = []cli.Container{
		{
			ID:        "a1b2c3d4e5f6",
			ShortID:   "a1b2c3d4",
			Image:     "ghcr.io/acme/api:1.4.2",
			Status:    "running",
			StartedAt: now.Add(-48 * time.Hour),
			Uptime:    48 * time.Hour,
			OS:        "linux",
			Arch:      "arm64",
			Networks: []cli.NetworkAttachment{
				{Network: "bridge", Hostname: "api-server", IP: "172.17.0.2"},
			},
			Ports: []cli.PortMapping{
				{HostPort: 8080, ContainerPort: 8080, Proto: "tcp"},
			},
			CPU:      4,
			MemBytes: 2 * 1024 * 1024 * 1024, // 2 GB
		},
		{
			ID:        "f6e5d4c3b2a1",
			ShortID:   "f6e5d4c3",
			Image:     "ghcr.io/acme/worker:2.1.0",
			Status:    "running",
			StartedAt: now.Add(-72 * time.Hour),
			Uptime:    72 * time.Hour,
			OS:        "linux",
			Arch:      "arm64",
			Networks: []cli.NetworkAttachment{
				{Network: "bridge", Hostname: "worker-1", IP: "172.17.0.3"},
			},
			CPU:      2,
			MemBytes: 1024 * 1024 * 1024, // 1 GB
		},
		{
			ID:        "123456789abc",
			ShortID:   "12345678",
			Image:     "ghcr.io/acme/worker:2.1.0",
			Status:    "running",
			StartedAt: now.Add(-36 * time.Hour),
			Uptime:    36 * time.Hour,
			OS:        "linux",
			Arch:      "arm64",
			Networks: []cli.NetworkAttachment{
				{Network: "bridge", Hostname: "worker-2", IP: "172.17.0.4"},
			},
			CPU:      2,
			MemBytes: 1024 * 1024 * 1024,
		},
		{
			ID:        "fedcba987654",
			ShortID:   "fedcba98",
			Image:     "docker.io/redis:7.2-alpine",
			Status:    "running",
			StartedAt: now.Add(-96 * time.Hour),
			Uptime:    96 * time.Hour,
			OS:        "linux",
			Arch:      "arm64",
			Networks: []cli.NetworkAttachment{
				{Network: "bridge", Hostname: "redis", IP: "172.17.0.5"},
			},
			Ports: []cli.PortMapping{
				{HostPort: 6379, ContainerPort: 6379, Proto: "tcp"},
			},
			CPU:      1,
			MemBytes: 512 * 1024 * 1024,
		},
		{
			ID:        "aabbccddeeff",
			ShortID:   "aabbccdd",
			Image:     "docker.io/nginx:1.25-alpine",
			Status:    "running",
			StartedAt: now.Add(-120 * time.Hour),
			Uptime:    120 * time.Hour,
			OS:        "linux",
			Arch:      "arm64",
			Networks: []cli.NetworkAttachment{
				{Network: "bridge", Hostname: "nginx-edge", IP: "172.17.0.6"},
			},
			Ports: []cli.PortMapping{
				{HostPort: 80, ContainerPort: 80, Proto: "tcp"},
				{HostPort: 443, ContainerPort: 443, Proto: "tcp"},
			},
			CPU:      2,
			MemBytes: 256 * 1024 * 1024,
		},
		{
			ID:        "112233445566",
			ShortID:   "11223344",
			Image:     "docker.io/postgres:16-alpine",
			Status:    "running",
			StartedAt: now.Add(-168 * time.Hour),
			Uptime:    168 * time.Hour,
			OS:        "linux",
			Arch:      "arm64",
			Networks: []cli.NetworkAttachment{
				{Network: "backend", Hostname: "postgres-db", IP: "192.168.1.10"},
			},
			Ports: []cli.PortMapping{
				{HostPort: 5432, ContainerPort: 5432, Proto: "tcp"},
			},
			CPU:      4,
			MemBytes: 4 * 1024 * 1024 * 1024,
		},
		{
			ID:        "998877665544",
			ShortID:   "99887766",
			Image:     "ghcr.io/acme/analytics:0.9.5",
			Status:    "running",
			StartedAt: now.Add(-24 * time.Hour),
			Uptime:    24 * time.Hour,
			OS:        "linux",
			Arch:      "amd64",
			Networks: []cli.NetworkAttachment{
				{Network: "backend", Hostname: "analytics", IP: "192.168.1.11"},
			},
			CPU:      8,
			MemBytes: 8 * 1024 * 1024 * 1024,
		},
		{
			ID:        "deadbeef0001",
			ShortID:   "deadbeef",
			Image:     "docker.io/grafana/grafana:10.3.0",
			Status:    "running",
			StartedAt: now.Add(-240 * time.Hour),
			Uptime:    240 * time.Hour,
			OS:        "linux",
			Arch:      "arm64",
			Networks: []cli.NetworkAttachment{
				{Network: "monitoring", Hostname: "grafana", IP: "10.0.0.10"},
			},
			Ports: []cli.PortMapping{
				{HostPort: 3000, ContainerPort: 3000, Proto: "tcp"},
			},
			CPU:      2,
			MemBytes: 1024 * 1024 * 1024,
		},
		// 3 stopped containers
		{
			ID:        "stopped000001",
			ShortID:   "stopped0",
			Image:     "ghcr.io/acme/migrator:1.0.0",
			Status:    "stopped",
			StartedAt: now.Add(-600 * time.Hour),
			Uptime:    0,
			OS:        "linux",
			Arch:      "arm64",
			CPU:       1,
			MemBytes:  256 * 1024 * 1024,
		},
		{
			ID:        "stopped000002",
			ShortID:   "stopped1",
			Image:     "docker.io/busybox:latest",
			Status:    "stopped",
			StartedAt: now.Add(-720 * time.Hour),
			Uptime:    0,
			OS:        "linux",
			Arch:      "amd64",
			CPU:       1,
			MemBytes:  64 * 1024 * 1024,
		},
		{
			ID:        "stopped000003",
			ShortID:   "stopped2",
			Image:     "ghcr.io/acme/backup:3.2.1",
			Status:    "stopped",
			StartedAt: now.Add(-48 * time.Hour),
			Uptime:    0,
			OS:        "linux",
			Arch:      "arm64",
			CPU:       2,
			MemBytes:  512 * 1024 * 1024,
		},
		// 1 stopped (pause may be unsupported, so set to stopped)
		{
			ID:        "stopped000004",
			ShortID:   "stopped3",
			Image:     "docker.io/alpine:3.19",
			Status:    "stopped",
			StartedAt: now.Add(-2 * time.Hour),
			Uptime:    0,
			OS:        "linux",
			Arch:      "arm64",
			CPU:       1,
			MemBytes:  128 * 1024 * 1024,
		},
	}

	// InspectContainer: return minimal JSON for first container
	fake.InspectContainerResp = []byte(`{"ID":"a1b2c3d4e5f6","Image":"ghcr.io/acme/api:1.4.2","Status":"running"}`)

	// --- 8 images across 3-4 repos ---
	fake.ListImagesResp = []cli.Image{
		{
			ID:         "img-api-142",
			ShortID:    "img-api-1",
			Repository: "ghcr.io/acme/api",
			Tag:        "1.4.2",
			Reference:  "ghcr.io/acme/api:1.4.2",
			Created:    now.Add(-30 * 24 * time.Hour),
			SizeBytes:  512 * 1024 * 1024,
		},
		{
			ID:         "img-worker-210",
			ShortID:    "img-worke",
			Repository: "ghcr.io/acme/worker",
			Tag:        "2.1.0",
			Reference:  "ghcr.io/acme/worker:2.1.0",
			Created:    now.Add(-45 * 24 * time.Hour),
			SizeBytes:  384 * 1024 * 1024,
		},
		{
			ID:         "img-redis-72",
			ShortID:    "img-redis",
			Repository: "docker.io/redis",
			Tag:        "7.2-alpine",
			Reference:  "docker.io/redis:7.2-alpine",
			Created:    now.Add(-60 * 24 * time.Hour),
			SizeBytes:  128 * 1024 * 1024,
		},
		{
			ID:         "img-nginx-125",
			ShortID:    "img-nginx",
			Repository: "docker.io/nginx",
			Tag:        "1.25-alpine",
			Reference:  "docker.io/nginx:1.25-alpine",
			Created:    now.Add(-90 * 24 * time.Hour),
			SizeBytes:  96 * 1024 * 1024,
		},
		{
			ID:         "img-postgres-16",
			ShortID:    "img-postg",
			Repository: "docker.io/postgres",
			Tag:        "16-alpine",
			Reference:  "docker.io/postgres:16-alpine",
			Created:    now.Add(-120 * 24 * time.Hour),
			SizeBytes:  256 * 1024 * 1024,
		},
		{
			ID:         "img-analytics-095",
			ShortID:    "img-analy",
			Repository: "ghcr.io/acme/analytics",
			Tag:        "0.9.5",
			Reference:  "ghcr.io/acme/analytics:0.9.5",
			Created:    now.Add(-15 * 24 * time.Hour),
			SizeBytes:  768 * 1024 * 1024,
		},
		{
			ID:         "img-grafana-1030",
			ShortID:    "img-grafa",
			Repository: "docker.io/grafana/grafana",
			Tag:        "10.3.0",
			Reference:  "docker.io/grafana/grafana:10.3.0",
			Created:    now.Add(-180 * 24 * time.Hour),
			SizeBytes:  320 * 1024 * 1024,
		},
		{
			ID:         "img-busybox-latest",
			ShortID:    "img-busyb",
			Repository: "docker.io/busybox",
			Tag:        "latest",
			Reference:  "docker.io/busybox:latest",
			Created:    now.Add(-365 * 24 * time.Hour),
			SizeBytes:  4 * 1024 * 1024,
		},
	}

	// InspectImage: return minimal JSON for first image
	fake.InspectImageResp = []byte(`{"ID":"img-api-142","Repository":"ghcr.io/acme/api","Tag":"1.4.2"}`)

	// --- 4 volumes ---
	fake.ListVolumesResp = []cli.Volume{
		{
			Name:       "api-data",
			Driver:     "local",
			Mountpoint: "/var/lib/container/volumes/api-data/_data",
			SizeBytes:  1024 * 1024 * 1024, // 1 GB
			UsedBy:     []string{"a1b2c3d4e5f6"},
		},
		{
			Name:       "postgres-data",
			Driver:     "local",
			Mountpoint: "/var/lib/container/volumes/postgres-data/_data",
			SizeBytes:  8 * 1024 * 1024 * 1024, // 8 GB
			UsedBy:     []string{"112233445566"},
		},
		{
			Name:       "redis-data",
			Driver:     "local",
			Mountpoint: "/var/lib/container/volumes/redis-data/_data",
			SizeBytes:  256 * 1024 * 1024,
			UsedBy:     []string{"fedcba987654"},
		},
		{
			Name:       "shared-logs",
			Driver:     "local",
			Mountpoint: "/var/lib/container/volumes/shared-logs/_data",
			SizeBytes:  512 * 1024 * 1024,
			UsedBy:     []string{"a1b2c3d4e5f6", "f6e5d4c3b2a1", "123456789abc"},
		},
	}

	// --- 3 networks ---
	fake.ListNetworksResp = []cli.Network{
		{
			Name:       "bridge",
			Driver:     "bridge",
			Subnet:     "172.17.0.0/16",
			Containers: []string{"a1b2c3d4e5f6", "f6e5d4c3b2a1", "123456789abc", "fedcba987654", "aabbccddeeff"},
		},
		{
			Name:       "backend",
			Driver:     "bridge",
			Subnet:     "192.168.1.0/24",
			Containers: []string{"112233445566", "998877665544"},
		},
		{
			Name:       "monitoring",
			Driver:     "bridge",
			Subnet:     "10.0.0.0/24",
			Containers: []string{"deadbeef0001"},
		},
	}

	// --- Builder running ---
	fake.BuilderStatusResp = cli.BuilderStatus{
		State:       "running",
		CPUs:        8,
		MemoryBytes: 16 * 1024 * 1024 * 1024, // 16 GB
		UptimeSec:   86400 * 7,               // 7 days
	}

	// --- 2 registries: one default ---
	fake.ListRegistriesResp = []cli.RegistryEntry{
		{
			Host:    "ghcr.io",
			User:    "acme-bot",
			Default: true,
		},
		{
			Host:    "docker.io",
			User:    "demo-user",
			Default: false,
		},
	}

	// --- 1 DNS domain (default) ---
	fake.ListDNSDomainsResp = []cli.DNSDomain{
		{
			Name:    "c9s.local",
			Default: true,
		},
	}

	// --- 5 system services ---
	fake.ListSystemServicesResp = []cli.SystemService{
		{
			Name:      "container-daemon",
			State:     "running",
			PID:       1234,
			UptimeSec: 86400 * 30, // 30 days
		},
		{
			Name:      "buildkitd",
			State:     "running",
			PID:       1235,
			UptimeSec: 86400 * 7, // 7 days
		},
		{
			Name:      "registry-proxy",
			State:     "running",
			PID:       1236,
			UptimeSec: 86400 * 14, // 14 days
		},
		{
			Name:      "network-controller",
			State:     "running",
			PID:       1237,
			UptimeSec: 86400 * 30,
		},
		{
			Name:      "volume-manager",
			State:     "running",
			PID:       1238,
			UptimeSec: 86400 * 30,
		},
	}

	// --- StreamLogs: 15-20 scripted log lines with timestamps ---
	fake.ReplayLogStream([]cli.StreamEvent{
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  Starting API server v1.4.2", now.Add(-10*time.Second).Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  Loaded configuration from /etc/api/config.yaml", now.Add(-9*time.Second).Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  Database connection pool initialized (max=50)", now.Add(-8*time.Second).Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  Connecting to Redis at redis:6379", now.Add(-7*time.Second).Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  Redis connection established", now.Add(-6*time.Second).Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  HTTP server listening on :8080", now.Add(-5*time.Second).Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  gRPC server listening on :9090", now.Add(-5*time.Second).Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  Health check endpoint ready at /health", now.Add(-4*time.Second).Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s WARN  Rate limiter threshold set to 1000 req/min", now.Add(-3*time.Second).Format(time.RFC3339)), Level: "WARN"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  Metrics endpoint exposed at /metrics", now.Add(-3*time.Second).Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  Request received: GET /api/users", now.Add(-2*time.Second).Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  Response sent: 200 OK (45ms)", now.Add(-2*time.Second).Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s WARN  Slow query detected: SELECT * FROM users (128ms)", now.Add(-1*time.Second).Format(time.RFC3339)), Level: "WARN"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  Background job scheduler started", now.Add(-1*time.Second).Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s ERROR Failed to send notification: connection timeout", now.Format(time.RFC3339)), Level: "ERROR"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  Retrying notification send...", now.Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  Notification sent successfully", now.Format(time.RFC3339)), Level: "INFO"},
		cli.LogLine{Raw: fmt.Sprintf("%s INFO  Cache hit ratio: 87%%", now.Format(time.RFC3339)), Level: "INFO"},
	}, 0)

	// --- StreamBuild: 6 scripted build steps ---
	fake.ReplayBuildStream([]cli.StreamEvent{
		cli.BuildStepEvent{Index: 1, Stage: "internal", Step: "load build definition from Containerfile", Duration: "0.1s", Status: "done"},
		cli.BuildStepEvent{Index: 2, Stage: "internal", Step: "load metadata for docker.io/library/golang:1.22-alpine", Duration: "0.5s", Status: "done"},
		cli.BuildStepEvent{Index: 3, Stage: "stage 1/2", Step: "FROM golang:1.22-alpine", Duration: "", Status: "cached"},
		cli.BuildStepEvent{Index: 4, Stage: "stage 1/2", Step: "COPY . /app", Duration: "1.2s", Status: "done"},
		cli.BuildStepEvent{Index: 5, Stage: "stage 1/2", Step: "RUN go build -o /app/server", Duration: "8.4s", Status: "done"},
		cli.BuildStepEvent{Index: 6, Stage: "stage 2/2", Step: "FROM alpine:3.19", Duration: "", Status: "cached"},
		cli.BuildStepEvent{Index: 7, Stage: "stage 2/2", Step: "COPY --from=0 /app/server /server", Duration: "0.3s", Status: "done"},
		cli.BuildStepEvent{Index: 8, Stage: "exporting", Step: "exporting to image", Duration: "1.1s", Status: "done"},
	}, 0)

	// --- StreamPull: 3 layers progressing ---
	fake.ReplayLayerStream([]cli.StreamEvent{
		cli.LayerProgressEvent{Layers: []cli.LayerProgress{
			{Digest: "sha256:abc123", State: "waiting", BytesDone: 0, BytesTotal: 0},
			{Digest: "sha256:def456", State: "waiting", BytesDone: 0, BytesTotal: 0},
			{Digest: "sha256:789ghi", State: "waiting", BytesDone: 0, BytesTotal: 0},
		}},
		cli.LayerProgressEvent{Layers: []cli.LayerProgress{
			{Digest: "sha256:abc123", State: "downloading", BytesDone: 1024 * 512, BytesTotal: 1024 * 1024},
			{Digest: "sha256:def456", State: "downloading", BytesDone: 1024 * 256, BytesTotal: 1024 * 512},
			{Digest: "sha256:789ghi", State: "waiting", BytesDone: 0, BytesTotal: 0},
		}},
		cli.LayerProgressEvent{Layers: []cli.LayerProgress{
			{Digest: "sha256:abc123", State: "extracting", BytesDone: 1024 * 1024, BytesTotal: 1024 * 1024},
			{Digest: "sha256:def456", State: "downloading", BytesDone: 1024 * 512, BytesTotal: 1024 * 512},
			{Digest: "sha256:789ghi", State: "downloading", BytesDone: 1024 * 128, BytesTotal: 1024 * 256},
		}},
		cli.LayerProgressEvent{Layers: []cli.LayerProgress{
			{Digest: "sha256:abc123", State: "done", BytesDone: 1024 * 1024, BytesTotal: 1024 * 1024},
			{Digest: "sha256:def456", State: "extracting", BytesDone: 1024 * 512, BytesTotal: 1024 * 512},
			{Digest: "sha256:789ghi", State: "downloading", BytesDone: 1024 * 256, BytesTotal: 1024 * 256},
		}},
		cli.LayerProgressEvent{Layers: []cli.LayerProgress{
			{Digest: "sha256:abc123", State: "done", BytesDone: 1024 * 1024, BytesTotal: 1024 * 1024},
			{Digest: "sha256:def456", State: "done", BytesDone: 1024 * 512, BytesTotal: 1024 * 512},
			{Digest: "sha256:789ghi", State: "extracting", BytesDone: 1024 * 256, BytesTotal: 1024 * 256},
		}},
		cli.LayerProgressEvent{Layers: []cli.LayerProgress{
			{Digest: "sha256:abc123", State: "done", BytesDone: 1024 * 1024, BytesTotal: 1024 * 1024},
			{Digest: "sha256:def456", State: "done", BytesDone: 1024 * 512, BytesTotal: 1024 * 512},
			{Digest: "sha256:789ghi", State: "done", BytesDone: 1024 * 256, BytesTotal: 1024 * 256},
		}},
	}, 0)

	return fake
}
