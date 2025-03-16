package queue

import (
	"sync"

	"github.com/desertwitch/gover/internal/generic/schema"
)

type EnumerationManager struct {
	sync.Mutex
	queues map[*EnumerationQueue]struct{}
}

func NewEnumerationManager() *EnumerationManager {
	return &EnumerationManager{
		queues: make(map[*EnumerationQueue]struct{}),
	}
}

func (e *EnumerationManager) GetItems() []*schema.Moveable {
	e.Lock()
	defer e.Unlock()

	result := []*schema.Moveable{}

	for q := range e.queues {
		result = append(result, q.GetItems()...)
	}

	return result
}

func (e *EnumerationManager) NewQueue() *EnumerationQueue {
	e.Lock()
	defer e.Unlock()

	q := NewEnumerationQueue()
	e.queues[q] = struct{}{}

	return q
}

func (e *EnumerationManager) DestroyQueue(q *EnumerationQueue) {
	e.Lock()
	defer e.Unlock()

	if _, exists := e.queues[q]; exists {
		delete(e.queues, q)
	}
}

func (e *EnumerationManager) DestroyQueues() {
	e.Lock()
	defer e.Unlock()

	e.queues = make(map[*EnumerationQueue]struct{})
}
