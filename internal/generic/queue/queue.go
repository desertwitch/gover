package queue

import (
	"sync"

	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/storage"
)

type Manager struct {
	sync.RWMutex
	items map[string]*DestinationQueue
}

func NewManager() *Manager {
	return &Manager{
		items: make(map[string]*DestinationQueue),
	}
}

func (b *Manager) Enqueue(items ...*filesystem.Moveable) {
	b.Lock()
	defer b.Unlock()

	for _, item := range items {
		if b.items[item.Dest.GetName()] == nil {
			b.items[item.Dest.GetName()] = NewDestinationQueue()
		}
		b.items[item.Dest.GetName()].Enqueue(item)
	}
}

func (b *Manager) Dequeue(target storage.Storage) (*filesystem.Moveable, bool) {
	b.Lock()
	defer b.Unlock()

	if queue, exists := b.items[target.GetName()]; exists {
		return queue.Dequeue()
	}

	return nil, false
}

func (b *Manager) GetQueueUnsafe(target storage.Storage) (*DestinationQueue, bool) {
	b.RLock()
	defer b.RUnlock()

	if queue, exists := b.items[target.GetName()]; exists {
		return queue, true
	}

	return nil, false
}

func (b *Manager) GetQueuesUnsafe() map[string]*DestinationQueue {
	b.RLock()
	defer b.RUnlock()

	return b.items
}

type DestinationQueue struct {
	sync.RWMutex
	head       int
	items      []*filesystem.Moveable
	inProgress map[*filesystem.Moveable]*TransferInfo
	success    map[*filesystem.Moveable]*TransferInfo
	skipped    map[*filesystem.Moveable]*TransferInfo
}

func NewDestinationQueue() *DestinationQueue {
	return &DestinationQueue{
		items:      []*filesystem.Moveable{},
		inProgress: make(map[*filesystem.Moveable]*TransferInfo),
		success:    make(map[*filesystem.Moveable]*TransferInfo),
		skipped:    make(map[*filesystem.Moveable]*TransferInfo),
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

func (q *DestinationQueue) SetSuccess(item *filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	transferInfo := q.inProgress[item]
	delete(q.inProgress, item)
	q.success[item] = transferInfo
}

func (q *DestinationQueue) SetSkipped(item *filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	transferInfo := q.inProgress[item]
	delete(q.inProgress, item)
	q.skipped[item] = transferInfo
}

func (q *DestinationQueue) SetProcessing(item *filesystem.Moveable) *TransferInfo {
	q.Lock()
	defer q.Unlock()

	transferInfo := &TransferInfo{}
	q.inProgress[item] = transferInfo

	return transferInfo
}
