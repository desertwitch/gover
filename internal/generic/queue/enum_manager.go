package queue

import (
	"sync"

	"github.com/desertwitch/gover/internal/generic/schema"
)

type EnumerationManager struct {
	sync.RWMutex
	queues map[string]*EnumerationShareQueue
}

func NewEnumerationManager() *EnumerationManager {
	return &EnumerationManager{
		queues: make(map[string]*EnumerationShareQueue),
	}
}

func (b *EnumerationManager) GetSuccessful() []*schema.Moveable {
	b.Lock()
	defer b.Unlock()

	result := []*schema.Moveable{}

	for _, q := range b.queues {
		result = append(result, q.GetSuccessful()...)
	}

	return result
}

func (b *EnumerationManager) Enqueue(items ...*schema.Moveable) {
	b.Lock()
	defer b.Unlock()

	for _, item := range items {
		if b.queues[item.Share.GetName()] == nil {
			b.queues[item.Share.GetName()] = NewEnumerationQueue()
		}
		b.queues[item.Share.GetName()].Enqueue(item)
	}
}

func (b *EnumerationManager) GetQueue(target schema.Share) (*EnumerationShareQueue, bool) {
	b.RLock()
	defer b.RUnlock()

	if queue, exists := b.queues[target.GetName()]; exists {
		return queue, true
	}

	return nil, false
}

// GetQueues returns a copy of the internal map holding pointers to all queues.
func (b *EnumerationManager) GetQueues() map[string]*EnumerationShareQueue {
	b.RLock()
	defer b.RUnlock()

	if b.queues == nil {
		return nil
	}

	queues := make(map[string]*EnumerationShareQueue)

	for k, v := range b.queues {
		queues[k] = v
	}

	return queues
}
