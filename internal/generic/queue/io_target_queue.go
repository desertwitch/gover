package queue

import (
	"context"
	"sync"

	"github.com/desertwitch/gover/internal/generic/schema"
)

type IOTargetQueue struct {
	sync.Mutex
	head       int
	items      []*schema.Moveable
	success    []*schema.Moveable
	skipped    []*schema.Moveable
	inProgress map[*schema.Moveable]struct{}
}

func NewIOTargetQueue() *IOTargetQueue {
	return &IOTargetQueue{
		items:      []*schema.Moveable{},
		inProgress: make(map[*schema.Moveable]struct{}),
		success:    []*schema.Moveable{},
		skipped:    []*schema.Moveable{},
	}
}

func (q *IOTargetQueue) HasRemainingItems() bool {
	q.Lock()
	defer q.Unlock()

	if q.head >= len(q.items) {
		return false
	}

	return true
}

func (q *IOTargetQueue) GetSuccessful() []*schema.Moveable {
	q.Lock()
	defer q.Unlock()

	result := make([]*schema.Moveable, len(q.success))
	copy(result, q.success)

	return result
}

func (q *IOTargetQueue) Enqueue(items ...*schema.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		if _, exists := q.inProgress[item]; exists {
			delete(q.inProgress, item)
		}
		q.items = append(q.items, item)
	}
}

func (q *IOTargetQueue) Dequeue() (*schema.Moveable, bool) {
	q.Lock()
	defer q.Unlock()

	if q.head >= len(q.items) {
		return nil, false
	}

	item := q.items[q.head]
	q.head++

	return item, true
}

func (q *IOTargetQueue) SetSuccess(items ...*schema.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.success = append(q.success, item)
	}
}

func (q *IOTargetQueue) SetSkipped(items ...*schema.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.skipped = append(q.skipped, item)
	}
}

func (q *IOTargetQueue) SetProcessing(items ...*schema.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		q.inProgress[item] = struct{}{}
	}
}

func (q *IOTargetQueue) DequeueAndProcess(ctx context.Context, processFunc func(*schema.Moveable) int) error {
	return processQueue(ctx, q, processFunc)
}

func (q *IOTargetQueue) DequeueAndProcessConc(ctx context.Context, maxWorkers int, processFunc func(*schema.Moveable) int) error {
	return concurrentProcessQueue(ctx, maxWorkers, q, processFunc)
}
