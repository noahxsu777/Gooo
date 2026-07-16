package queue

import (
	"testing"
	"time"

	"github.com/noahxsu777/Gooo/internal/events"
)

func msg(id string, p events.Priority) events.Message {
	return events.Message{ID: id, Priority: p, Timestamp: time.Now()}
}

func TestBoundedQueueBackpressure(t *testing.T) {
	q := New(2)
	if !q.Enqueue(msg("1", events.PriorityLow)) || !q.Enqueue(msg("2", events.PriorityLow)) {
		t.Fatal("enqueue failed")
	}
	if q.Enqueue(msg("3", events.PriorityLow)) {
		t.Fatal("expected drop when full with same priority")
	}
	if !q.Enqueue(msg("4", events.PriorityHigh)) {
		t.Fatal("high priority should replace low")
	}
}
