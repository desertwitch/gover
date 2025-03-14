package queue

import (
	"sync"

	"github.com/desertwitch/gover/internal/generic/filesystem"
)

type IOTargetQueue struct {
	sync.Mutex
	head       int
	items      []*filesystem.Moveable
	success    []*filesystem.Moveable
	skipped    []*filesystem.Moveable
	inProgress map[*filesystem.Moveable]struct{}
}

func NewIOTargetQueue() *IOTargetQueue {
	return &IOTargetQueue{
		items:      []*filesystem.Moveable{},
		inProgress: make(map[*filesystem.Moveable]struct{}),
		success:    []*filesystem.Moveable{},
		skipped:    []*filesystem.Moveable{},
	}
}

func (q *IOTargetQueue) ResetQueue() {
	q.Lock()
	defer q.Unlock()

	q.items = make([]*filesystem.Moveable, len(q.success))
	copy(q.items, q.success)

	q.head = 0
	q.success = []*filesystem.Moveable{}
	q.skipped = []*filesystem.Moveable{}
	q.inProgress = make(map[*filesystem.Moveable]struct{})
}

func (q *IOTargetQueue) Enqueue(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	q.items = append(q.items, items...)
}

func (q *IOTargetQueue) Dequeue() (*filesystem.Moveable, bool) {
	q.Lock()
	defer q.Unlock()

	if q.head >= len(q.items) {
		return nil, false
	}

	item := q.items[q.head]
	q.head++

	return item, true
}

func (q *IOTargetQueue) SetSuccess(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.success = append(q.success, item)
	}
}

func (q *IOTargetQueue) SetSkipped(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.skipped = append(q.skipped, item)
	}
}

func (q *IOTargetQueue) SetProcessing(items ...*filesystem.Moveable) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		q.inProgress[item] = struct{}{}
	}
}
