// Package clock provides a small abstraction over time so tests can
// drive ticks deterministically. Production code uses Real(); tests use NewFake().
package clock

import (
	"sync"
	"time"
)

// Clock is the surface used by the rest of c9s.
type Clock interface {
	Now() time.Time
	// Tick returns a channel that receives the current (fake or real) time
	// approximately every d. The channel is buffered (size 1) so a slow
	// consumer doesn't block the producer.
	Tick(d time.Duration) <-chan time.Time
}

// Real returns a Clock backed by the standard library.
func Real() Clock { return realClock{} }

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// Tick returns a single-shot channel that fires once after d. Each call
// allocates a fresh time.After timer, which the runtime garbage-collects
// after firing.
//
// Note: callers re-arm by invoking Tick again (e.g. inside a tea.Cmd that
// returns a TickMsg). This is fine because every Tick is one-shot and
// the previous goroutine/timer has already exited; there is no
// accumulation. Earlier versions of this method spawned a long-lived
// goroutine + time.Ticker per call, which leaked one goroutine + ticker
// per refresh — a 2-second cadence over an hour produced ~1800 leaks
// per screen. See PR review C2 of v0.1.0 for the symptom.
func (realClock) Tick(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// Fake is a manually-driven clock for tests.
type Fake struct {
	mu      sync.Mutex
	now     time.Time
	tickers []*fakeTicker
}

type fakeTicker struct {
	interval time.Duration
	next     time.Time
	ch       chan time.Time
}

// NewFake constructs a Fake at the given start time.
func NewFake(start time.Time) *Fake { return &Fake{now: start} }

// Now returns the current fake time.
func (f *Fake) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

// Tick returns a channel; it fires once for every multiple of d that
// the fake clock has advanced past.
func (f *Fake) Tick(d time.Duration) <-chan time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	ch := make(chan time.Time, 1)
	f.tickers = append(f.tickers, &fakeTicker{
		interval: d,
		next:     f.now.Add(d),
		ch:       ch,
	})
	return ch
}

// Advance moves the fake clock forward by d, firing any tickers whose
// next time falls within the new range.
func (f *Fake) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	target := f.now.Add(d)
	for _, t := range f.tickers {
		for !t.next.After(target) {
			select {
			case t.ch <- t.next:
			default:
			}
			t.next = t.next.Add(t.interval)
		}
	}
	f.now = target
}
