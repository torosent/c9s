package clock_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/torosent/c9s/internal/clock"
)

// TestRealTick_NoGoroutineLeak is a regression test for the C2 bug from
// the v0.1.0 review: clock.Real().Tick(d) used to spawn a goroutine that
// looped forever forwarding ticks from a never-stopped time.Ticker, and
// the goroutine was unreachable so callers couldn't stop it. Calling Tick
// 1000 times leaked ~1000 goroutines + tickers.
//
// The fix makes Tick a thin wrapper around time.After (single-shot),
// which the runtime cleans up after firing. Asserting on goroutine
// counts is inherently racy (background runtime goroutines come and
// go), so we tolerate a small amount of slop.
func TestRealTick_NoGoroutineLeak(t *testing.T) {
	clk := clock.Real()

	// Drive a few ticks first to warm up the runtime so the baseline is
	// stable.
	for i := 0; i < 5; i++ {
		<-clk.Tick(time.Microsecond)
	}
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	const N = 200
	for i := 0; i < N; i++ {
		<-clk.Tick(time.Microsecond)
	}
	// Give any remaining goroutines a chance to wind down before we GC.
	time.Sleep(50 * time.Millisecond)
	runtime.GC()
	time.Sleep(50 * time.Millisecond)

	got := runtime.NumGoroutine()
	// With the old buggy implementation, got would be roughly baseline + N
	// (200+ extra goroutines). With the fix it's typically baseline +/- 1
	// because each time.After timer is GC'd after firing.
	if got > baseline+10 {
		t.Errorf("expected goroutine count near %d, got %d (delta=%d, leaked?)", baseline, got, got-baseline)
	}
}
