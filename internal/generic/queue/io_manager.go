package queue

import (
	"sync"

	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/schema"
)

type IOManager struct {
	sync.RWMutex
	queues map[string]*IOTargetQueue
}

func NewIOManager() *IOManager {
	return &IOManager{
		queues: make(map[string]*IOTargetQueue),
	}
}

func (b *IOManager) Enqueue(items ...*filesystem.Moveable) {
	b.Lock()
	defer b.Unlock()

	for _, item := range items {
		if b.queues[item.Dest.GetName()] == nil {
			b.queues[item.Dest.GetName()] = NewIOTargetQueue()
		}
		b.queues[item.Dest.GetName()].Enqueue(item)
	}
}

func (b *IOManager) GetQueueUnsafe(target schema.Storage) (*IOTargetQueue, bool) {
	b.RLock()
	defer b.RUnlock()

	if queue, exists := b.queues[target.GetName()]; exists {
		return queue, true
	}

	return nil, false
}

func (b *IOManager) GetQueuesUnsafe() map[string]*IOTargetQueue {
	b.RLock()
	defer b.RUnlock()

	return b.queues
}
