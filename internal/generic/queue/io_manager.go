package queue

import (
	"sync"

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

func (b *IOManager) Enqueue(items ...*schema.Moveable) {
	b.Lock()
	defer b.Unlock()

	for _, item := range items {
		if b.queues[item.Dest.GetName()] == nil {
			b.queues[item.Dest.GetName()] = NewIOTargetQueue()
		}
		b.queues[item.Dest.GetName()].Enqueue(item)
	}
}

func (b *IOManager) GetQueue(target schema.Storage) (*IOTargetQueue, bool) {
	b.RLock()
	defer b.RUnlock()

	if queue, exists := b.queues[target.GetName()]; exists {
		return queue, true
	}

	return nil, false
}

// GetQueues returns a copy of the internal map holding pointers to all queues.
func (b *IOManager) GetQueues() map[string]*IOTargetQueue {
	b.RLock()
	defer b.RUnlock()

	if b.queues == nil {
		return nil
	}

	queues := make(map[string]*IOTargetQueue)

	for k, v := range b.queues {
		queues[k] = v
	}

	return queues
}
