package events

import (
	"testing"
	"time"
)

func TestDeduper(t *testing.T) {
	d := NewDeduper(time.Minute)
	now := time.Now()
	d.nowFunc = func() time.Time { return now }
	if d.IsDuplicate("a") {
		t.Fatal("first event should not be duplicate")
	}
	if !d.IsDuplicate("a") {
		t.Fatal("second event should be duplicate")
	}
	now = now.Add(2 * time.Minute)
	if d.IsDuplicate("a") {
		t.Fatal("event should expire by ttl")
	}
}
