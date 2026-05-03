package demodata_test

import (
	"context"
	"testing"
	"time"

	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/cli/demodata"
	"github.com/torosent/c9s/internal/clock"
)

func TestNewFake_Counts(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC))
	fake := demodata.NewFake(clk)

	ctx := context.Background()

	// Verify container count
	containers, err := fake.ListContainers(ctx, false)
	if err != nil {
		t.Fatalf("ListContainers: %v", err)
	}
	if got, want := len(containers), 12; got != want {
		t.Errorf("containers: got %d, want %d", got, want)
	}

	// Verify 8 running + rest stopped/paused
	runningCount := 0
	for _, c := range containers {
		if c.Status == "running" {
			runningCount++
		}
	}
	if runningCount != 8 {
		t.Errorf("running containers: got %d, want 8", runningCount)
	}

	// Verify image count
	images, err := fake.ListImages(ctx)
	if err != nil {
		t.Fatalf("ListImages: %v", err)
	}
	if got, want := len(images), 8; got != want {
		t.Errorf("images: got %d, want %d", got, want)
	}

	// Verify volume count
	volumes, err := fake.ListVolumes(ctx)
	if err != nil {
		t.Fatalf("ListVolumes: %v", err)
	}
	if got, want := len(volumes), 4; got != want {
		t.Errorf("volumes: got %d, want %d", got, want)
	}

	// Verify network count
	networks, err := fake.ListNetworks(ctx)
	if err != nil {
		t.Fatalf("ListNetworks: %v", err)
	}
	if got, want := len(networks), 3; got != want {
		t.Errorf("networks: got %d, want %d", got, want)
	}

	// Verify builder
	builder, err := fake.BuilderStatus(ctx)
	if err != nil {
		t.Fatalf("BuilderStatus: %v", err)
	}
	if builder.State != "running" {
		t.Errorf("builder state: got %q, want running", builder.State)
	}

	// Verify registries
	registries, err := fake.ListRegistries(ctx)
	if err != nil {
		t.Fatalf("ListRegistries: %v", err)
	}
	if got, want := len(registries), 2; got != want {
		t.Errorf("registries: got %d, want %d", got, want)
	}
	hasDefault := false
	for _, r := range registries {
		if r.Default {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		t.Error("no default registry found")
	}

	// Verify DNS domains
	domains, err := fake.ListDNSDomains(ctx)
	if err != nil {
		t.Fatalf("ListDNSDomains: %v", err)
	}
	if got, want := len(domains), 1; got != want {
		t.Errorf("dns domains: got %d, want %d", got, want)
	}

	// Verify system services
	services, err := fake.ListSystemServices(ctx)
	if err != nil {
		t.Fatalf("ListSystemServices: %v", err)
	}
	if got, want := len(services), 5; got != want {
		t.Errorf("system services: got %d, want %d", got, want)
	}
}

func TestNewFake_Roundtrip(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC))
	fake := demodata.NewFake(clk)
	ctx := context.Background()

	// Get a container
	containers, err := fake.ListContainers(ctx, false)
	if err != nil {
		t.Fatalf("ListContainers: %v", err)
	}
	if len(containers) == 0 {
		t.Fatal("no containers")
	}

	cID := containers[0].ID

	// Inspect it
	inspectJSON, err := fake.InspectContainer(ctx, cID)
	if err != nil {
		t.Fatalf("InspectContainer: %v", err)
	}
	if len(inspectJSON) == 0 {
		t.Error("InspectContainer returned empty JSON")
	}

	// Stream logs
	stream, err := fake.StreamLogs(ctx, cID, false)
	if err != nil {
		t.Fatalf("StreamLogs: %v", err)
	}

	var logEvents []cli.StreamEvent
	for ev := range stream.Events {
		logEvents = append(logEvents, ev)
	}
	result := <-stream.Done

	if result.Err != nil {
		t.Errorf("StreamLogs error: %v", result.Err)
	}
	if len(logEvents) == 0 {
		t.Error("no log events")
	}
	// Verify we got some log lines (15-20 expected)
	logLineCount := 0
	for _, ev := range logEvents {
		if _, ok := ev.(cli.LogLine); ok {
			logLineCount++
		}
	}
	if logLineCount < 10 {
		t.Errorf("expected at least 10 log lines, got %d", logLineCount)
	}
}

func TestNewFake_StreamBuild(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC))
	fake := demodata.NewFake(clk)
	ctx := context.Background()

	stream, err := fake.StreamBuild(ctx, cli.BuildOpts{
		ContextPath: ".",
		Tag:         "demo:latest",
	})
	if err != nil {
		t.Fatalf("StreamBuild: %v", err)
	}

	var buildEvents []cli.StreamEvent
	for ev := range stream.Events {
		buildEvents = append(buildEvents, ev)
	}
	result := <-stream.Done

	if result.Err != nil {
		t.Errorf("StreamBuild error: %v", result.Err)
	}
	if len(buildEvents) == 0 {
		t.Error("no build events")
	}

	// Verify we got some build step events (6 expected)
	stepCount := 0
	for _, ev := range buildEvents {
		if _, ok := ev.(cli.BuildStepEvent); ok {
			stepCount++
		}
	}
	if stepCount < 5 {
		t.Errorf("expected at least 5 build steps, got %d", stepCount)
	}
}

func TestNewFake_StreamPull(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC))
	fake := demodata.NewFake(clk)
	ctx := context.Background()

	stream, err := fake.StreamPull(ctx, "demo:latest")
	if err != nil {
		t.Fatalf("StreamPull: %v", err)
	}

	var pullEvents []cli.StreamEvent
	for ev := range stream.Events {
		pullEvents = append(pullEvents, ev)
	}
	result := <-stream.Done

	if result.Err != nil {
		t.Errorf("StreamPull error: %v", result.Err)
	}
	if len(pullEvents) == 0 {
		t.Error("no pull events")
	}

	// Verify we got layer progress events
	layerCount := 0
	for _, ev := range pullEvents {
		if _, ok := ev.(cli.LayerProgressEvent); ok {
			layerCount++
		}
	}
	if layerCount < 3 {
		t.Errorf("expected at least 3 layer progress events, got %d", layerCount)
	}
}
