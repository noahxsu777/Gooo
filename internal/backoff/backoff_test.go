package backoff

import (
	"testing"
	"time"
)

func TestDurationWithCap(t *testing.T) {
	b := Exponential{Base: time.Second, Max: 5 * time.Second, Jitter: 0}
	if got := b.Duration(0); got != time.Second {
		t.Fatalf("expected 1s, got %v", got)
	}
	if got := b.Duration(4); got != 5*time.Second {
		t.Fatalf("expected cap 5s, got %v", got)
	}
}

func TestCanRetry(t *testing.T) {
	b := Exponential{Retries: 3}
	if !b.CanRetry(2) {
		t.Fatal("attempt 2 should retry")
	}
	if b.CanRetry(3) {
		t.Fatal("attempt 3 should stop")
	}
}
