package queue

import (
	"sync"

	"github.com/desertwitch/gover/internal/generic/schema"
)

type EnumerationManager struct {
	sync.RWMutex
	queues map[string]*EnumerationSourceQueue
}

func NewEnumerationManager() *EnumerationManager {
	return &EnumerationManager{
		queues: make(map[string]*EnumerationSourceQueue),
	}
}

func (b *EnumerationManager) Enqueue(items ...*EnumerationTask) {
	b.Lock()
	defer b.Unlock()

	for _, item := range items {
		if b.queues[item.Source.GetName()] == nil {
			b.queues[item.Source.GetName()] = NewEnumerationSourceQueue()
		}
		b.queues[item.Source.GetName()].Enqueue(item)
	}
}

func (b *EnumerationManager) GetQueue(source schema.Storage) (*EnumerationSourceQueue, bool) {
	b.RLock()
	defer b.RUnlock()

	if queue, exists := b.queues[source.GetName()]; exists {
		return queue, true
	}

	return nil, false
}

// GetQueues returns a copy of the internal map holding pointers to all queues.
func (b *EnumerationManager) GetQueues() map[string]*EnumerationSourceQueue {
	b.RLock()
	defer b.RUnlock()

	if b.queues == nil {
		return nil
	}

	queues := make(map[string]*EnumerationSourceQueue)

	for k, v := range b.queues {
		queues[k] = v
	}

	return queues
}
