package queue

import (
	"sync"

	"github.com/desertwitch/gover/internal/generic/schema"
)

type EnumerationManager struct {
	sync.Mutex
	queues map[*EnumerationQueue]string // map[*EnumerationQueue]queueStage
}

func NewEnumerationManager() *EnumerationManager {
	return &EnumerationManager{
		queues: make(map[*EnumerationQueue]string),
	}
}

func (e *EnumerationManager) GetSuccessful() []*schema.Moveable {
	e.Lock()
	defer e.Unlock()

	result := []*schema.Moveable{}

	for q := range e.queues {
		result = append(result, q.GetSuccessful()...)
	}

	return result
}

func (e *EnumerationManager) NewQueue() *EnumerationQueue {
	e.Lock()
	defer e.Unlock()

	q := NewEnumerationQueue()
	e.queues[q] = ""

	return q
}

func (e *EnumerationManager) SetQueuePhase(q *EnumerationQueue, phase string) {
	e.Lock()
	defer e.Unlock()

	if _, exists := e.queues[q]; exists {
		e.queues[q] = phase
	}
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

	e.queues = make(map[*EnumerationQueue]string)
}
