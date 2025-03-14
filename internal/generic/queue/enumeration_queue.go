package queue

import (
	"context"
	"sync"

	"github.com/desertwitch/gover/internal/generic/schema"
)

type EnumerationQueue struct {
	sync.Mutex
	head       int
	items      []*schema.Moveable
	success    []*schema.Moveable
	skipped    []*schema.Moveable
	inProgress map[*schema.Moveable]struct{}
}

func NewEnumerationQueue() *EnumerationQueue {
	return &EnumerationQueue{
		head:       0,
		items:      []*schema.Moveable{},
		success:    []*schema.Moveable{},
		skipped:    []*schema.Moveable{},
		inProgress: make(map[*schema.Moveable]struct{}),
	}
}

func (q *EnumerationQueue) ResetQueue() {
	q.Lock()
	defer q.Unlock()

	q.items = make([]*schema.Moveable, len(q.success))
	copy(q.items, q.success)

	q.head = 0
	q.success = []*schema.Moveable{}
	q.skipped = []*schema.Moveable{}
	q.inProgress = make(map[*schema.Moveable]struct{})
}

func (q *EnumerationQueue) GetItems() []*schema.Moveable {
	q.Lock()
	defer q.Unlock()

	result := make([]*schema.Moveable, len(q.items))
	copy(result, q.items)

	return result
}

func (q *EnumerationQueue) Enqueue(items ...*schema.Moveable) {
	q.Lock()
	defer q.Unlock()

	q.items = append(q.items, items...)
}

func (q *EnumerationQueue) Dequeue() (*schema.Moveable, bool) {
	q.Lock()
	defer q.Unlock()

	if q.head >= len(q.items) {
		return nil, false
	}

	item := q.items[q.head]
	q.head++

	return item, true
}

func (q *EnumerationQueue) SetSuccess(items ...*schema.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.success = append(q.success, item)
	}
}

func (q *EnumerationQueue) SetSkipped(items ...*schema.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.skipped = append(q.skipped, item)
	}
}

func (q *EnumerationQueue) SetProcessing(items ...*schema.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		q.inProgress[item] = struct{}{}
	}
}

func (q *EnumerationQueue) DequeueAndProcess(ctx context.Context, processFunc func(*schema.Moveable) bool, resetQueueAfter bool) error {
	return processQueue(ctx, q, processFunc, resetQueueAfter)
}

func (q *EnumerationQueue) DequeueAndProcessConc(ctx context.Context, maxWorkers int, processFunc func(*schema.Moveable) bool, resetQueueAfter bool) error {
	return concurrentProcessQueue(ctx, maxWorkers, q, processFunc, resetQueueAfter)
}
