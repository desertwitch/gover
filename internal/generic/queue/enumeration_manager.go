package queue

import (
	"sync"

	"github.com/desertwitch/gover/internal/generic/filesystem"
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

func (e *EnumerationManager) GetItems() []*filesystem.Moveable {
	e.Lock()
	defer e.Unlock()

	files := []*filesystem.Moveable{}

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
