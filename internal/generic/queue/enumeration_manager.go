package queue

import (
	"sync"

	"github.com/desertwitch/gover/internal/generic/schema"
)

type EnumerationManager struct {
	sync.Mutex
	queues []*EnumerationQueue
}

func NewEnumerationManager() *EnumerationManager {
	return &EnumerationManager{
		queues: []*EnumerationQueue{},
	}
}

func (e *EnumerationManager) GetItems() []*schema.Moveable {
	e.Lock()
	defer e.Unlock()

	files := []*schema.Moveable{}

	for _, q := range e.queues {
		files = append(files, q.GetItems()...)
	}

	return files
}

func (e *EnumerationManager) NewQueue() *EnumerationQueue {
	e.Lock()
	defer e.Unlock()

	q := NewEnumerationQueue()
	e.queues = append(e.queues, q)

	return q
}
