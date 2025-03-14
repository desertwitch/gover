package queue

import (
	"sync"

	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/storage"
)

type IOManager struct {
	sync.RWMutex
	queues map[string]*DestinationQueue
}

func NewIOManager() *IOManager {
	return &IOManager{
		queues: make(map[string]*DestinationQueue),
	}
}

func (b *IOManager) Enqueue(items ...*filesystem.Moveable) {
	b.Lock()
	defer b.Unlock()

	for _, item := range items {
		if b.queues[item.Dest.GetName()] == nil {
			b.queues[item.Dest.GetName()] = NewDestinationQueue()
		}
		b.queues[item.Dest.GetName()].Enqueue(item)
	}
}

func (b *IOManager) GetQueueUnsafe(target storage.Storage) (*DestinationQueue, bool) {
	b.RLock()
	defer b.RUnlock()

	if queue, exists := b.queues[target.GetName()]; exists {
		return queue, true
	}

	return nil, false
}

func (b *IOManager) GetQueuesUnsafe() map[string]*DestinationQueue {
	b.RLock()
	defer b.RUnlock()

	return b.queues
}

type DestinationQueue struct {
	sync.Mutex
	head       int
	items      []*filesystem.Moveable
	success    []*filesystem.Moveable
	skipped    []*filesystem.Moveable
	inProgress map[*filesystem.Moveable]struct{}
}

func NewDestinationQueue() *DestinationQueue {
	return &DestinationQueue{
		items:      []*filesystem.Moveable{},
		inProgress: make(map[*filesystem.Moveable]struct{}),
		success:    []*filesystem.Moveable{},
		skipped:    []*filesystem.Moveable{},
	}
}

func (q *DestinationQueue) ResetQueue() {
	q.Lock()
	defer q.Unlock()

	q.items = make([]*filesystem.Moveable, len(q.success))
	copy(q.items, q.success)

	q.head = 0
	q.success = []*filesystem.Moveable{}
	q.skipped = []*filesystem.Moveable{}
	q.inProgress = make(map[*filesystem.Moveable]struct{})
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
		delete(q.inProgress, item)
		q.success = append(q.success, item)
	}
}

func (q *DestinationQueue) SetSkipped(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.skipped = append(q.skipped, item)
	}
}

func (q *DestinationQueue) SetProcessing(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		q.inProgress[item] = struct{}{}
	}
}
