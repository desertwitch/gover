package queue

import (
	"context"
	"sync"

	"github.com/desertwitch/gover/internal/generic/schema"
)

type EnumerationTask struct {
	Share    schema.Share
	Source   schema.Storage
	Function func() error
}

func (e *EnumerationTask) Run() error {
	return e.Function()
}

type EnumerationShareQueue struct {
	sync.Mutex
	head       int
	items      []*EnumerationTask
	success    []*EnumerationTask
	skipped    []*EnumerationTask
	inProgress map[*EnumerationTask]struct{}
}

func NewEnumerationShareQueue() *EnumerationShareQueue {
	return &EnumerationShareQueue{
		items:      []*EnumerationTask{},
		inProgress: make(map[*EnumerationTask]struct{}),
		success:    []*EnumerationTask{},
		skipped:    []*EnumerationTask{},
	}
}

func (q *EnumerationShareQueue) HasRemainingItems() bool {
	q.Lock()
	defer q.Unlock()

	if q.head >= len(q.items) {
		return false
	}

	return true
}

func (q *EnumerationShareQueue) GetSuccessful() []*EnumerationTask {
	q.Lock()
	defer q.Unlock()

	result := make([]*EnumerationTask, len(q.success))
	copy(result, q.success)

	return result
}

func (q *EnumerationShareQueue) Enqueue(items ...*EnumerationTask) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		if _, exists := q.inProgress[item]; exists {
			delete(q.inProgress, item)
		}
		q.items = append(q.items, item)
	}
}

func (q *EnumerationShareQueue) Dequeue() (*EnumerationTask, bool) {
	q.Lock()
	defer q.Unlock()

	if q.head >= len(q.items) {
		return nil, false
	}

	item := q.items[q.head]
	q.head++

	return item, true
}

func (q *EnumerationShareQueue) SetSuccess(items ...*EnumerationTask) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.success = append(q.success, item)
	}
}

func (q *EnumerationShareQueue) SetSkipped(items ...*EnumerationTask) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.skipped = append(q.skipped, item)
	}
}

func (q *EnumerationShareQueue) SetProcessing(items ...*EnumerationTask) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		q.inProgress[item] = struct{}{}
	}
}

func (q *EnumerationShareQueue) DequeueAndProcess(ctx context.Context, processFunc func(*EnumerationTask) int) error {
	return processQueue(ctx, q, processFunc)
}

func (q *EnumerationShareQueue) DequeueAndProcessConc(ctx context.Context, maxWorkers int, processFunc func(*EnumerationTask) int) error {
	return concurrentProcessQueue(ctx, maxWorkers, q, processFunc)
}
