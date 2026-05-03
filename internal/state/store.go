// Package state holds the snapshot caches that screens consume.
// Plan 2 will grow this with refresh helpers; v0.1.0 only provides
// the generic Store + Snapshot.
package state

import (
	"sync"
	"time"
)

// Snapshot is one observation of a resource list: the items, when it
// was fetched, and any error that occurred during fetch.
type Snapshot[T any] struct {
	Items     []T
	FetchedAt time.Time
	Err       error
}

// Store is a goroutine-safe holder for a single Snapshot[T].
// Each resource type gets its own Store (e.g. Store[Container],
// Store[Image]).
type Store[T any] struct {
	mu   sync.RWMutex
	snap Snapshot[T]
}

// NewStore returns an empty Store.
func NewStore[T any]() *Store[T] { return &Store[T]{} }

// Set replaces the current snapshot.
func (s *Store[T]) Set(snap Snapshot[T]) {
	s.mu.Lock()
	s.snap = snap
	s.mu.Unlock()
}

// Get returns the current snapshot. The Items slice is shared; callers
// that mutate must copy first.
func (s *Store[T]) Get() Snapshot[T] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snap
}
