package cli

import (
	"testing"
	"time"
)

// drainStream consumes Events until the channel is closed, then reads the
// final StreamResult. Tests must use this instead of a select{} on Events
// and Done together — those race because Events is buffered and the producer
// closes Events before sending on Done; select picks randomly between a
// ready Events delivery and a ready Done delivery, sometimes returning before
// all events have been observed.
func drainStream(t *testing.T, s Stream) ([]StreamEvent, StreamResult) {
	t.Helper()
	var events []StreamEvent
	for ev := range s.Events {
		events = append(events, ev)
	}
	select {
	case res := <-s.Done:
		return events, res
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for stream Done after Events closed")
		return events, StreamResult{}
	}
}
