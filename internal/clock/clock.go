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

func (realClock) Tick(d time.Duration) <-chan time.Time {
	out := make(chan time.Time, 1)
	go func() {
		t := time.NewTicker(d)
		defer t.Stop()
		for tick := range t.C {
			select {
			case out <- tick:
			default: // drop if consumer is behind
			}
		}
	}()
	return out
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
