package events

import (
	"sync"
	"time"
)

type Deduper struct {
	mu      sync.Mutex
	seen    map[string]time.Time
	ttl     time.Duration
	nowFunc func() time.Time
}

func NewDeduper(ttl time.Duration) *Deduper {
	return &Deduper{
		seen:    make(map[string]time.Time),
		ttl:     ttl,
		nowFunc: time.Now,
	}
}

func (d *Deduper) IsDuplicate(id string) bool {
	if id == "" {
		return false
	}
	now := d.nowFunc()
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cleanup(now)
	if expiry, ok := d.seen[id]; ok && expiry.After(now) {
		return true
	}
	d.seen[id] = now.Add(d.ttl)
	return false
}

func (d *Deduper) cleanup(now time.Time) {
	for id, expiry := range d.seen {
		if !expiry.After(now) {
			delete(d.seen, id)
		}
	}
}
