package webhook

import (
	"sync"
	"time"
)

type fixedWindow struct {
	count int
	start time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	clients map[string]fixedWindow
	nowFunc func() time.Time
}

func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		max:     max,
		window:  window,
		clients: map[string]fixedWindow{},
		nowFunc: time.Now,
	}
}

func (r *RateLimiter) Allow(key string) bool {
	now := r.nowFunc()
	r.mu.Lock()
	defer r.mu.Unlock()
	entry := r.clients[key]
	if entry.start.IsZero() || now.Sub(entry.start) > r.window {
		entry = fixedWindow{count: 0, start: now}
	}
	entry.count++
	r.clients[key] = entry
	return entry.count <= r.max
}
