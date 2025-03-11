package queue

import (
	"sync"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
)

type Manager struct {
	sync.RWMutex
	items map[unraid.Storeable]*DestinationQueue
}

func NewManager() *Manager {
	return &Manager{
		items: make(map[unraid.Storeable]*DestinationQueue),
	}
}

func (b *Manager) Enqueue(items ...*filesystem.Moveable) {
	b.Lock()
	defer b.Unlock()

	for _, item := range items {
		if b.items[item.Dest] == nil {
			b.items[item.Dest] = NewDestinationQueue()
		}
		b.items[item.Dest].Enqueue(item)
	}
}

func (b *Manager) Dequeue(target unraid.Storeable) (*filesystem.Moveable, bool) {
	b.Lock()
	defer b.Unlock()

	if queue, exists := b.items[target]; exists {
		return queue.Dequeue()
	}

	return nil, false
}

func (b *Manager) GetQueueUnsafe(target unraid.Storeable) (*DestinationQueue, bool) {
	b.RLock()
	defer b.RUnlock()

	if queue, exists := b.items[target]; exists {
		return queue, true
	}

	return nil, false
}

func (b *Manager) GetQueuesUnsafe() map[unraid.Storeable]*DestinationQueue {
	b.RLock()
	defer b.RUnlock()

	return b.items
}

type DestinationQueue struct {
	sync.Mutex
	head       int
	items      []*filesystem.Moveable
	Success    []*filesystem.Moveable
	Skipped    []*filesystem.Moveable
	InProgress map[*filesystem.Moveable]struct{}
}

func NewDestinationQueue() *DestinationQueue {
	return &DestinationQueue{
		items:      []*filesystem.Moveable{},
		InProgress: make(map[*filesystem.Moveable]struct{}),
		Success:    []*filesystem.Moveable{},
		Skipped:    []*filesystem.Moveable{},
	}
}

func (q *DestinationQueue) Enqueue(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	q.items = append(q.items, items...)
}

func (q *DestinationQueue) Dequeue() (*filesystem.Moveable, bool) {
	q.Lock()
	defer q.Unlock()

	if q.head >= len(q.items) {
		return nil, false
	}

	item := q.items[q.head]
	q.head++

	return item, true
}

func (q *DestinationQueue) SetSuccess(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.InProgress, item)
		q.Success = append(q.Success, item)
	}
}

func (q *DestinationQueue) SetSkipped(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.InProgress, item)
		q.Skipped = append(q.Skipped, item)
	}
}

func (q *DestinationQueue) SetProcessing(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		q.InProgress[item] = struct{}{}
	}
}
