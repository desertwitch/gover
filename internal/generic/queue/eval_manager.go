package queue

import (
	"sync"

	"github.com/desertwitch/gover/internal/generic/schema"
)

type EvaluationManager struct {
	sync.RWMutex
	queues map[string]*EvaluationShareQueue
}

func NewEvaluationManager() *EvaluationManager {
	return &EvaluationManager{
		queues: make(map[string]*EvaluationShareQueue),
	}
}

func (b *EvaluationManager) GetSuccessful() []*schema.Moveable {
	b.Lock()
	defer b.Unlock()

	result := []*schema.Moveable{}

	for _, q := range b.queues {
		result = append(result, q.GetSuccessful()...)
	}

	return result
}

func (b *EvaluationManager) Enqueue(items ...*schema.Moveable) {
	b.Lock()
	defer b.Unlock()

	for _, item := range items {
		if b.queues[item.Share.GetName()] == nil {
			b.queues[item.Share.GetName()] = NewEvaluationShareQueue()
		}
		b.queues[item.Share.GetName()].Enqueue(item)
	}
}

func (b *EvaluationManager) GetQueue(target schema.Share) (*EvaluationShareQueue, bool) {
	b.RLock()
	defer b.RUnlock()

	if queue, exists := b.queues[target.GetName()]; exists {
		return queue, true
	}

	return nil, false
}

// GetQueues returns a copy of the internal map holding pointers to all queues.
func (b *EvaluationManager) GetQueues() map[string]*EvaluationShareQueue {
	b.RLock()
	defer b.RUnlock()

	if b.queues == nil {
		return nil
	}

	queues := make(map[string]*EvaluationShareQueue)

	for k, v := range b.queues {
		queues[k] = v
	}

	return queues
}
