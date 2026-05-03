package jobs

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
)

func TestManager_Submit(t *testing.T) {
	clk := clock.NewFake(time.Now())
	mgr := New(clk)

	fake := cli.NewFake()
	events := []cli.StreamEvent{
		cli.LogLine{Raw: "line 1", Level: "INFO"},
		cli.LogLine{Raw: "line 2", Level: ""},
	}
	fake.ReplayLogStream(events, 0)

	stream, err := fake.StreamLogs(context.Background(), "test-id", false)
	if err != nil {
		t.Fatalf("StreamLogs failed: %v", err)
	}

	job := mgr.Submit(KindLog, "test-id", stream)
	jobID := job.ID

	if jobID == "" {
		t.Error("expected non-empty job ID")
	}

	// Get a safe copy of the job
	job, ok := mgr.Get(jobID)
	if !ok {
		t.Fatal("job not found")
	}

	if job.Kind != KindLog {
		t.Errorf("kind = %q, want %q", job.Kind, KindLog)
	}
	if job.Target != "test-id" {
		t.Errorf("target = %q, want test-id", job.Target)
	}
	if job.State != StateRunning && job.State != StateDone {
		t.Errorf("state = %q, want running or done", job.State)
	}

	// Wait for job to complete
	time.Sleep(100 * time.Millisecond)

	job, ok = mgr.Get(jobID)
	if !ok {
		t.Fatal("job not found")
	}

	if job.State != StateDone {
		t.Errorf("state = %q, want %q", job.State, StateDone)
	}
	if len(job.Lines) != 2 {
		t.Errorf("lines count = %d, want 2", len(job.Lines))
	}
	if job.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", job.ExitCode)
	}
}

func TestManager_Cancel(t *testing.T) {
	clk := clock.NewFake(time.Now())
	mgr := New(clk)

	fake := cli.NewFake()
	// Create a stream with some events
	events := []cli.StreamEvent{
		cli.RawLine{Text: "line 1"},
		cli.RawLine{Text: "line 2"},
	}
	fake.ReplayLogStream(events, 0)

	stream, err := fake.StreamLogs(context.Background(), "test-id", true)
	if err != nil {
		t.Fatalf("StreamLogs failed: %v", err)
	}

	job := mgr.Submit(KindLog, "test-id", stream)

	// Cancel immediately (before stream completes)
	if err := mgr.Cancel(job.ID); err != nil {
		t.Errorf("Cancel failed: %v", err)
	}

	// Wait for cancellation to propagate
	time.Sleep(100 * time.Millisecond)

	job, ok := mgr.Get(job.ID)
	if !ok {
		t.Fatal("job not found")
	}

	// Job should be either cancelled or done (race condition)
	// The important thing is Cancel() didn't error
	if job.State != StateCancelled && job.State != StateDone && job.State != StateFailed {
		t.Errorf("unexpected state = %q", job.State)
	}
}

func TestManager_List(t *testing.T) {
	clk := clock.NewFake(time.Now())
	mgr := New(clk)

	fake := cli.NewFake()
	fake.ReplayLogStream([]cli.StreamEvent{cli.RawLine{Text: "a"}}, 0)
	stream1, _ := fake.StreamLogs(context.Background(), "id1", false)
	job1 := mgr.Submit(KindLog, "id1", stream1)

	fake.ReplayBuildStream([]cli.StreamEvent{cli.RawLine{Text: "b"}}, 0)
	stream2, _ := fake.StreamBuild(context.Background(), cli.BuildOpts{ContextPath: "/test"})
	job2 := mgr.Submit(KindBuild, "/test", stream2)

	time.Sleep(100 * time.Millisecond)

	jobs := mgr.List()
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	ids := []string{jobs[0].ID, jobs[1].ID}
	if !contains(ids, job1.ID) || !contains(ids, job2.ID) {
		t.Errorf("expected jobs %s and %s, got %v", job1.ID, job2.ID, ids)
	}
}

func TestManager_Get(t *testing.T) {
	clk := clock.NewFake(time.Now())
	mgr := New(clk)

	fake := cli.NewFake()
	fake.ReplayLogStream([]cli.StreamEvent{}, 0)
	stream, _ := fake.StreamLogs(context.Background(), "test", false)
	job := mgr.Submit(KindLog, "test", stream)

	got, ok := mgr.Get(job.ID)
	if !ok {
		t.Fatal("job not found")
	}
	if got.ID != job.ID {
		t.Errorf("got job ID %s, want %s", got.ID, job.ID)
	}

	_, ok = mgr.Get("nonexistent")
	if ok {
		t.Error("expected false for nonexistent job")
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	// This test runs with -race to detect data races
	clk := clock.NewFake(time.Now())
	mgr := New(clk)

	var wg sync.WaitGroup
	const goroutines = 10

	// Submit jobs concurrently
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			fake := cli.NewFake()
			fake.ReplayLogStream([]cli.StreamEvent{
				cli.LogLine{Raw: "test", Level: ""},
			}, 0)
			stream, _ := fake.StreamLogs(context.Background(), "test", false)
			mgr.Submit(KindLog, "test", stream)
		}(i)
	}

	// List jobs concurrently while submitting
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = mgr.List()
		}()
	}

	wg.Wait()

	// Final check
	jobs := mgr.List()
	if len(jobs) != goroutines {
		t.Errorf("expected %d jobs, got %d", goroutines, len(jobs))
	}
}

func TestManager_LineAccumulation(t *testing.T) {
	clk := clock.NewFake(time.Now())
	mgr := New(clk)

	fake := cli.NewFake()
	events := []cli.StreamEvent{
		cli.LogLine{Raw: "line 1", Level: "INFO"},
		cli.LogLine{Raw: "line 2", Level: "WARN"},
		cli.LogLine{Raw: "line 3", Level: ""},
		cli.BuildStepEvent{Index: 1, Stage: "test", Step: "step 1"},
	}
	fake.ReplayLogStream(events, 0)

	stream, _ := fake.StreamLogs(context.Background(), "test", false)
	job := mgr.Submit(KindLog, "test", stream)

	time.Sleep(150 * time.Millisecond)

	job, ok := mgr.Get(job.ID)
	if !ok {
		t.Fatal("job not found")
	}

	if len(job.Lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(job.Lines))
	}

	// Check that lines contain expected content
	combined := strings.Join(job.Lines, "\n")
	if !strings.Contains(combined, "line 1") {
		t.Error("missing 'line 1' in accumulated lines")
	}
	if !strings.Contains(combined, "step 1") {
		t.Error("missing 'step 1' in accumulated lines")
	}
}

func TestJob_Elapsed(t *testing.T) {
	clk := clock.NewFake(time.Now())
	start := clk.Now()

	job := &Job{
		ID:      "test",
		Kind:    KindLog,
		Target:  "test",
		State:   StateRunning,
		Started: start,
	}

	clk.Advance(5 * time.Second)

	elapsed := job.Elapsed(clk)
	if elapsed != 5*time.Second {
		t.Errorf("elapsed = %v, want 5s", elapsed)
	}

	// Complete the job
	job.State = StateDone
	job.Ended = clk.Now()

	clk.Advance(10 * time.Second)

	// Elapsed should still be 5s (frozen at completion time)
	elapsed = job.Elapsed(clk)
	if elapsed != 5*time.Second {
		t.Errorf("elapsed = %v, want 5s (frozen)", elapsed)
	}
}

func contains(strs []string, s string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}
