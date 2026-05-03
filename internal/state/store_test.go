package state

import (
	"errors"
	"testing"
	"time"
)

func TestStoreEmpty(t *testing.T) {
	s := NewStore[int]()
	snap := s.Get()
	if snap.Items != nil {
		t.Errorf("Items = %v, want nil", snap.Items)
	}
	if !snap.FetchedAt.IsZero() {
		t.Errorf("FetchedAt = %v, want zero", snap.FetchedAt)
	}
	if snap.Err != nil {
		t.Errorf("Err = %v, want nil", snap.Err)
	}
}

func TestStoreSetAndGet(t *testing.T) {
	s := NewStore[string]()
	now := time.Date(2026, 5, 2, 9, 0, 0, 0, time.UTC)
	s.Set(Snapshot[string]{
		Items:     []string{"a", "b"},
		FetchedAt: now,
	})
	got := s.Get()
	if len(got.Items) != 2 || got.Items[0] != "a" || got.Items[1] != "b" {
		t.Errorf("Items = %v", got.Items)
	}
	if !got.FetchedAt.Equal(now) {
		t.Errorf("FetchedAt = %v", got.FetchedAt)
	}
}

func TestStoreSetReplaces(t *testing.T) {
	s := NewStore[int]()
	s.Set(Snapshot[int]{Items: []int{1, 2, 3}})
	s.Set(Snapshot[int]{Items: []int{4}, Err: errors.New("partial")})
	got := s.Get()
	if len(got.Items) != 1 || got.Items[0] != 4 {
		t.Errorf("Items = %v, want [4]", got.Items)
	}
	if got.Err == nil {
		t.Errorf("Err = nil, want non-nil")
	}
}
