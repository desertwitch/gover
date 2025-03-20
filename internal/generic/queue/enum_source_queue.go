package queue

import (
	"context"
	"sync"

	"github.com/desertwitch/gover/internal/generic/schema"
)

type EnumerationTask struct {
	Share    schema.Share
	Source   schema.Storage
	Function func() int
}

func (e *EnumerationTask) Run() int {
	return e.Function()
}

type EnumerationSourceQueue struct {
	sync.Mutex
	head       int
	items      []*EnumerationTask
	success    []*EnumerationTask
	skipped    []*EnumerationTask
	inProgress map[*EnumerationTask]struct{}
}

func NewEnumerationSourceQueue() *EnumerationSourceQueue {
	return &EnumerationSourceQueue{
		items:      []*EnumerationTask{},
		inProgress: make(map[*EnumerationTask]struct{}),
		success:    []*EnumerationTask{},
		skipped:    []*EnumerationTask{},
	}
}

func (q *EnumerationSourceQueue) HasRemainingItems() bool {
	q.Lock()
	defer q.Unlock()

	if q.head >= len(q.items) {
		return false
	}

	return true
}

func (q *EnumerationSourceQueue) GetSuccessful() []*EnumerationTask {
	q.Lock()
	defer q.Unlock()

	result := make([]*EnumerationTask, len(q.success))
	copy(result, q.success)

	return result
}

func (q *EnumerationSourceQueue) Enqueue(items ...*EnumerationTask) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.items = append(q.items, item)
	}
}

func (q *EnumerationSourceQueue) Dequeue() (*EnumerationTask, bool) {
	q.Lock()
	defer q.Unlock()

	if q.head >= len(q.items) {
		return nil, false
	}

	item := q.items[q.head]
	q.head++

	return item, true
}

func (q *EnumerationSourceQueue) SetSuccess(items ...*EnumerationTask) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.success = append(q.success, item)
	}
}

func (q *EnumerationSourceQueue) SetSkipped(items ...*EnumerationTask) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.skipped = append(q.skipped, item)
	}
}

func (q *EnumerationSourceQueue) SetProcessing(items ...*EnumerationTask) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		q.inProgress[item] = struct{}{}
	}
}

func (q *EnumerationSourceQueue) DequeueAndProcess(ctx context.Context, processFunc func(*EnumerationTask) int) error {
	return processQueue(ctx, q, processFunc)
}

func (q *EnumerationSourceQueue) DequeueAndProcessConc(ctx context.Context, maxWorkers int, processFunc func(*EnumerationTask) int) error {
	return concurrentProcessQueue(ctx, maxWorkers, q, processFunc)
}
