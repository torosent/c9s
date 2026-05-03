package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func makeFixtureScript(t *testing.T, fixture string) string {
	t.Helper()
	abs, _ := filepath.Abs(filepath.Join("testdata", fixture))
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	script := "#!/bin/bash\ncat " + abs + "\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return scriptPath
}

func TestListSystemServices(t *testing.T) {
	c := NewDefaultClient(WithBinary(makeFixtureScript(t, "system-services.json")))
	svcs, err := c.ListSystemServices(context.Background())
	if err != nil {
		t.Fatalf("ListSystemServices: %v", err)
	}
	if len(svcs) != 3 {
		t.Fatalf("expected 3 services, got %d", len(svcs))
	}
	if svcs[0].Name != "container-runtime" || svcs[0].State != "running" || svcs[0].PID != 1234 {
		t.Errorf("got %+v", svcs[0])
	}
	if svcs[2].State != "stopped" {
		t.Errorf("got %+v", svcs[2])
	}
}

func TestSystemDF(t *testing.T) {
	c := NewDefaultClient(WithBinary(makeFixtureScript(t, "system-df.json")))
	df, err := c.SystemDF(context.Background())
	if err != nil {
		t.Fatalf("SystemDF: %v", err)
	}
	if df.Images.Count != 12 || df.Images.SizeBytes != 1073741824 {
		t.Errorf("Images = %+v", df.Images)
	}
	if df.Containers.Count != 4 {
		t.Errorf("Containers = %+v", df.Containers)
	}
	if df.Volumes.ReclaimBytes != 104857600 {
		t.Errorf("Volumes = %+v", df.Volumes)
	}
}

func TestSystemDFNull(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho 'null'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	df, err := c.SystemDF(context.Background())
	if err != nil {
		t.Fatalf("SystemDF null: %v", err)
	}
	if df.Images.Count != 0 {
		t.Errorf("expected zero df, got %+v", df)
	}
}

func TestListDNSDomains(t *testing.T) {
	c := NewDefaultClient(WithBinary(makeFixtureScript(t, "system-dns.json")))
	doms, err := c.ListDNSDomains(context.Background())
	if err != nil {
		t.Fatalf("ListDNSDomains: %v", err)
	}
	if len(doms) != 2 {
		t.Fatalf("expected 2 domains")
	}
	if doms[0].Name != "internal.local" || !doms[0].Default {
		t.Errorf("got %+v", doms[0])
	}
}

func TestListSystemProperties(t *testing.T) {
	c := NewDefaultClient(WithBinary(makeFixtureScript(t, "system-properties.json")))
	props, err := c.ListSystemProperties(context.Background())
	if err != nil {
		t.Fatalf("ListSystemProperties: %v", err)
	}
	if len(props) != 3 {
		t.Fatalf("expected 3 properties")
	}
	if !props[2].ReadOnly {
		t.Errorf("expected version readonly")
	}
}

func TestSystemLifecycleVoidCalls(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	ctx := context.Background()
	for _, fn := range []func() error{
		func() error { return c.SystemStartAll(ctx) },
		func() error { return c.SystemStopAll(ctx) },
		func() error { return c.CreateDNSDomain(ctx, "x") },
		func() error { return c.DeleteDNSDomain(ctx, "x") },
		func() error { return c.SetDefaultDNSDomain(ctx, "x") },
		func() error { return c.SetSystemProperty(ctx, "k", "v") },
		func() error { return c.ResetSystemProperty(ctx, "k") },
	} {
		if err := fn(); err != nil {
			t.Errorf("call returned error: %v", err)
		}
	}
}

func TestStreamSystemLogs(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "container")
	script := "#!/bin/bash\necho 'INFO: started'\necho 'ERROR: nope'\nexit 0\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	c := NewDefaultClient(WithBinary(scriptPath))
	stream, err := c.StreamSystemLogs(context.Background(), false)
	if err != nil {
		t.Fatalf("StreamSystemLogs: %v", err)
	}
	gotEvents := 0
	for ev := range stream.Events {
		_ = ev
		gotEvents++
	}
	<-stream.Done
	if gotEvents == 0 {
		t.Errorf("expected events from system logs stream")
	}
}

func TestFakeSystem(t *testing.T) {
	f := NewFake()
	f.ListSystemServicesResp = []SystemService{{Name: "s1"}}
	f.SystemDFResp = SystemDF{Images: DFSection{Count: 1}}
	f.ListDNSDomainsResp = []DNSDomain{{Name: "d1"}}
	f.ListSystemPropertiesResp = []SystemProperty{{Key: "k"}}
	ctx := context.Background()
	_, _ = f.ListSystemServices(ctx)
	_, _ = f.SystemDF(ctx)
	_, _ = f.ListDNSDomains(ctx)
	_, _ = f.ListSystemProperties(ctx)
	_ = f.SystemStartAll(ctx)
	_ = f.SystemStopAll(ctx)
	_ = f.CreateDNSDomain(ctx, "x")
	_ = f.DeleteDNSDomain(ctx, "x")
	_ = f.SetDefaultDNSDomain(ctx, "x")
	_ = f.SetSystemProperty(ctx, "a", "b")
	_ = f.ResetSystemProperty(ctx, "a")
	stream, err := f.StreamSystemLogs(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	stream.Cancel()
	for range stream.Events {
	}
	<-stream.Done
	wantPrefix := []string{
		"ListSystemServices", "SystemDF", "ListDNSDomains", "ListSystemProperties",
		"SystemStartAll", "SystemStopAll", "CreateDNSDomain", "DeleteDNSDomain",
		"SetDefaultDNSDomain", "SetSystemProperty", "ResetSystemProperty",
	}
	for i, w := range wantPrefix {
		if i >= len(f.Calls) || f.Calls[i] != w {
			t.Errorf("Calls[%d] = %v, want %v", i, f.Calls, wantPrefix)
			break
		}
	}
}
