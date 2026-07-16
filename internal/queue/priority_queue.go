package queue

import (
	"container/heap"
	"sync"

	"github.com/noahxsu777/Gooo/internal/events"
)

type item struct {
	msg      events.Message
	priority int
	index    int
}

type heapItems []*item

func (h heapItems) Len() int { return len(h) }
func (h heapItems) Less(i, j int) bool {
	if h[i].priority == h[j].priority {
		return h[i].msg.Timestamp.Before(h[j].msg.Timestamp)
	}
	return h[i].priority > h[j].priority
}
func (h heapItems) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}
func (h *heapItems) Push(x any) {
	it := x.(*item)
	it.index = len(*h)
	*h = append(*h, it)
}
func (h *heapItems) Pop() any {
	old := *h
	n := len(old)
	it := old[n-1]
	*h = old[:n-1]
	return it
}

type PriorityQueue struct {
	mu      sync.Mutex
	items   heapItems
	maxSize int
	dropped int64
}

func New(maxSize int) *PriorityQueue {
	pq := &PriorityQueue{maxSize: maxSize}
	heap.Init(&pq.items)
	return pq
}

func (p *PriorityQueue) Enqueue(msg events.Message) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.maxSize > 0 && p.items.Len() >= p.maxSize {
		lowestIdx := p.lowestPriorityIndex()
		if lowestIdx >= 0 && int(msg.Priority) > p.items[lowestIdx].priority {
			heap.Remove(&p.items, lowestIdx)
		} else {
			p.dropped++
			return false
		}
	}
	heap.Push(&p.items, &item{msg: msg, priority: int(msg.Priority)})
	return true
}

func (p *PriorityQueue) Dequeue() (events.Message, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.items.Len() == 0 {
		return events.Message{}, false
	}
	it := heap.Pop(&p.items).(*item)
	return it.msg, true
}

func (p *PriorityQueue) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.items.Len()
}

func (p *PriorityQueue) Dropped() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.dropped
}

func (p *PriorityQueue) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.items = heapItems{}
	heap.Init(&p.items)
}

func (p *PriorityQueue) lowestPriorityIndex() int {
	if len(p.items) == 0 {
		return -1
	}
	idx := 0
	for i := 1; i < len(p.items); i++ {
		if p.items[i].priority < p.items[idx].priority {
			idx = i
		}
	}
	return idx
}
