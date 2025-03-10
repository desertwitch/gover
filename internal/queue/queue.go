package queue

import (
	"sync"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
)

type BucketQueue struct {
	sync.RWMutex
	items map[unraid.Storeable]*QueueManager
}

func NewBucketQueue() *BucketQueue {
	return &BucketQueue{
		items: make(map[unraid.Storeable]*QueueManager),
	}
}

func (b *BucketQueue) Enqueue(items ...*filesystem.Moveable) {
	b.Lock()
	defer b.Unlock()

	for _, item := range items {
		if b.items[item.Dest] == nil {
			b.items[item.Dest] = NewQueueManager()
		}
		b.items[item.Dest].Enqueue(item)
	}
}

func (b *BucketQueue) Dequeue(target unraid.Storeable) (*filesystem.Moveable, bool) {
	b.Lock()
	defer b.Unlock()

	if queue, exists := b.items[target]; exists {
		return queue.Dequeue()
	}

	return nil, false
}

func (b *BucketQueue) GetQueueUnsafe(target unraid.Storeable) (*QueueManager, bool) {
	b.RLock()
	defer b.RUnlock()

	if queue, exists := b.items[target]; exists {
		return queue, true
	}

	return nil, false
}

func (b *BucketQueue) GetQueuesUnsafe() map[unraid.Storeable]*QueueManager {
	b.RLock()
	defer b.RUnlock()

	return b.items
}

type QueueManager struct {
	sync.Mutex
	head       int
	items      []*filesystem.Moveable
	Success    []*filesystem.Moveable
	Skipped    []*filesystem.Moveable
	InProgress map[*filesystem.Moveable]struct{}
}

func NewQueueManager() *QueueManager {
	return &QueueManager{
		items:      []*filesystem.Moveable{},
		InProgress: make(map[*filesystem.Moveable]struct{}),
		Success:    []*filesystem.Moveable{},
		Skipped:    []*filesystem.Moveable{},
	}
}

func (q *QueueManager) Enqueue(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	q.items = append(q.items, items...)
}

func (q *QueueManager) Dequeue() (*filesystem.Moveable, bool) {
	q.Lock()
	defer q.Unlock()

	if q.head >= len(q.items) {
		return nil, false
	}

	item := q.items[q.head]
	q.head++

	return item, true
}

func (q *QueueManager) SetSuccess(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.InProgress, item)
		q.Success = append(q.Success, item)
	}
}

func (q *QueueManager) SetSkipped(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.InProgress, item)
		q.Skipped = append(q.Skipped, item)
	}
}

func (q *QueueManager) SetProcessing(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		q.InProgress[item] = struct{}{}
	}
}
