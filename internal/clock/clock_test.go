package clock

import (
	"testing"
	"time"
)

func TestRealClockNow(t *testing.T) {
	c := Real()
	before := time.Now()
	got := c.Now()
	after := time.Now()
	if got.Before(before) || got.After(after) {
		t.Fatalf("Real().Now() = %v, want between %v and %v", got, before, after)
	}
}

func TestFakeClockAdvance(t *testing.T) {
	start := time.Date(2026, 5, 2, 9, 0, 0, 0, time.UTC)
	c := NewFake(start)
	if got := c.Now(); !got.Equal(start) {
		t.Fatalf("Now() = %v, want %v", got, start)
	}
	c.Advance(2 * time.Second)
	if got := c.Now(); !got.Equal(start.Add(2 * time.Second)) {
		t.Fatalf("after Advance: Now() = %v, want %v", got, start.Add(2*time.Second))
	}
}

func TestFakeClockTickFiresOnAdvance(t *testing.T) {
	c := NewFake(time.Unix(0, 0))
	ch := c.Tick(time.Second)

	select {
	case <-ch:
		t.Fatal("tick fired before Advance")
	default:
	}

	c.Advance(time.Second)
	select {
	case <-ch:
		// ok
	case <-time.After(100 * time.Millisecond):
		t.Fatal("tick did not fire after Advance(1s)")
	}
}

func TestRealClockTickFires(t *testing.T) {
	c := Real()
	ch := c.Tick(20 * time.Millisecond)
	select {
	case <-ch:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("real ticker did not fire within 500ms")
	}
}
