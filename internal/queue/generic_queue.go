package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/desertwitch/gover/internal/schema"
)

const (
	// DecisionSuccess is returned by a processFunc when an item was processed.
	DecisionSuccess = 1

	// DecisionSkipped is returned by a processFunc when an item was skipped.
	DecisionSkipped = 0

	// DecisionRequeue is returned by a processFunc when an item needs
	// requeueing.
	DecisionRequeue = -1
)

// GenericQueue is a generic queue that can hold any comparable type of items.
type GenericQueue[V comparable] struct {
	sync.RWMutex

	hasStarted  bool
	hasFinished bool
	startTime   time.Time
	finishTime  time.Time
	head        int
	items       []V
	success     []V
	skipped     []V
	inProgress  map[V]struct{}
}

// NewGenericQueue returns a pointer to a new [GenericQueue].
func NewGenericQueue[V comparable]() *GenericQueue[V] {
	return &GenericQueue[V]{
		inProgress: make(map[V]struct{}),
	}
}

// HasRemainingItems returns whether a queue has remaining items to process.
func (q *GenericQueue[V]) HasRemainingItems() bool {
	q.RLock()
	defer q.RUnlock()

	if q.head >= len(q.items) {
		return false
	}

	return true
}

// GetSuccessful returns a copy of the internal slice holding all successful
// items.
func (q *GenericQueue[V]) GetSuccessful() []V {
	q.RLock()
	defer q.RUnlock()

	result := make([]V, len(q.success))
	copy(result, q.success)

	return result
}

// Enqueue adds items to the queue.
func (q *GenericQueue[V]) Enqueue(items ...V) {
	q.Lock()
	defer q.Unlock()

	if q.hasFinished {
		q.finishTime = time.Time{}
		q.hasFinished = false
	}

	for _, item := range items {
		delete(q.inProgress, item)
		q.items = append(q.items, item)
	}
}

// Dequeue returns an item from the queue and advances the queue head.
func (q *GenericQueue[V]) Dequeue() (V, bool) { //nolint:ireturn
	q.Lock()
	defer q.Unlock()

	if q.head >= len(q.items) {
		var zeroVal V

		return zeroVal, false
	}

	if q.head == len(q.items)-1 {
		if !q.hasFinished {
			q.finishTime = time.Now()
			q.hasFinished = true
		}
	}

	if !q.hasStarted {
		q.startTime = time.Now()
		q.hasStarted = true
	}

	item := q.items[q.head]
	q.head++

	return item, true
}

// SetSuccess sets given in-progress queue items as successfully processed. The
// items are removed from the in-progress map in the process.
func (q *GenericQueue[V]) SetSuccess(items ...V) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.success = append(q.success, item)
	}
}

// SetSkipped sets given in-progress queue items as skipped. The items are
// removed from the in-progress map in the process.
func (q *GenericQueue[V]) SetSkipped(items ...V) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.skipped = append(q.skipped, item)
	}
}

// SetProcessing sets given items as in progress (processing).
func (q *GenericQueue[V]) SetProcessing(items ...V) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		q.inProgress[item] = struct{}{}
	}
}

// Progress returns the [Progress] for the [GenericQueue].
func (q *GenericQueue[V]) Progress() Progress {
	q.RLock()
	defer q.RUnlock()

	hasStarted := q.hasStarted
	totalItems := len(q.items)

	processedItems := len(q.success) + len(q.skipped)
	processedItems = min(processedItems, totalItems)

	var progressPct float64
	if totalItems > 0 {
		progressPct = float64(processedItems) / float64(totalItems) * 100 //nolint:mnd
		progressPct = max(float64(0), min(progressPct, float64(100)))     //nolint:mnd
	}

	var eta time.Time
	var timeLeft time.Duration

	var transferSpeed float64
	transferSpeedUnit := "items/sec"

	if hasStarted && processedItems > 0 && processedItems < totalItems {
		elapsed := time.Since(q.startTime)
		itemsPerSec := float64(processedItems) / max(elapsed.Seconds(), 1)

		if itemsPerSec > 0 {
			remainingItems := totalItems - processedItems
			remainingSeconds := float64(remainingItems) / itemsPerSec
			timeLeft = time.Duration(remainingSeconds * float64(time.Second))
			eta = time.Now().Add(timeLeft)
			transferSpeed = itemsPerSec
		}
	}

	return Progress{
		HasStarted:        hasStarted,
		HasFinished:       q.hasFinished,
		StartTime:         q.startTime,
		FinishTime:        q.finishTime,
		ProgressPct:       progressPct,
		TotalItems:        totalItems,
		ProcessedItems:    processedItems,
		InProgressItems:   len(q.inProgress),
		SuccessItems:      len(q.success),
		SkippedItems:      len(q.skipped),
		ETA:               eta,
		TimeLeft:          timeLeft,
		TransferSpeed:     transferSpeed,
		TransferSpeedUnit: transferSpeedUnit,
	}
}

// DequeueAndProcess sequentially dequeues and processes items using the given
// processFunc. An error is only returned in case of a context cancellation, the
// processFunc is otherwise expected to return only an integer with the
// processing function's decision for that item.
//
// Possible decisions to be returned: [DecisionSuccess], [DecisionSkipped],
// [DecisionRequeue].
func (q *GenericQueue[V]) DequeueAndProcess(ctx context.Context, processFunc func(V) int) error {
	for ctx.Err() != nil {
		item, ok := q.Dequeue()
		if !ok {
			break
		}

		q.SetProcessing(item)

		switch processFunc(item) {
		case DecisionRequeue:
			q.Enqueue(item)

		case DecisionSkipped:
			q.SetSkipped(item)

		case DecisionSuccess:
			q.SetSuccess(item)
		}
	}

	if ctx.Err() != nil {
		return fmt.Errorf("(queue-proc) %w", ctx.Err())
	}

	return nil
}

// DequeueAndProcessConc concurrently dequeues and processes items using given
// processFunc. An error is only returned in case of a context cancellation, the
// processFunc is otherwise expected to return only an integer with the
// processing function's decision for that item.
//
// Possible decisions to be returned: [DecisionSuccess], [DecisionSkipped],
// [DecisionRequeue].
//
// It is the responsibility of the processFunc to ensure thread-safety for
// anything happening inside the processFunc, with the [GenericQueue] only
// guaranteeing thread-safety for itself.
func (q *GenericQueue[V]) DequeueAndProcessConc(ctx context.Context, maxWorkers int, processFunc func(V) int) error {
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, maxWorkers)

LOOP:
	for {
		select {
		case <-ctx.Done():
			wg.Wait()

			return fmt.Errorf("(queue-concproc) %w", ctx.Err())
		case semaphore <- struct{}{}:
		}

		item, ok := q.Dequeue()
		if !ok {
			<-semaphore

			break
		}

		wg.Add(1)
		go func(item V) {
			defer wg.Done()
			defer func() { <-semaphore }()

			q.SetProcessing(item)

			switch processFunc(item) {
			case DecisionRequeue:
				q.Enqueue(item)

			case DecisionSkipped:
				q.SetSkipped(item)

			case DecisionSuccess:
				q.SetSuccess(item)
			}
		}(item)
	}

	wg.Wait()

	if ctx.Err() != nil {
		return fmt.Errorf("(queue-concproc) %w", ctx.Err())
	}

	if q.HasRemainingItems() {
		// In case item(s) were requeued but all workers have already left.
		goto LOOP
	}

	return nil
}

// PreProcess runs a [schema.Pipeline]'s contained pre-processors on all yet
// unprocessed queue items.
//
// If the queue is being operated on concurrently, sorting functions should not
// be used as pre-processors. Instead, such functions should be post-processors
// instead to guarantee that the order is preserved at the end of the queue.
func (q *GenericQueue[V]) PreProcess(p schema.Pipeline[V]) bool {
	if items, ok := p.PreProcess(q.items); ok {
		q.items = items

		return true
	}

	return false
}

// PostProcess runs a [schema.Pipeline]'s contained post-processors on all
// successfully processed queue items.
func (q *GenericQueue[V]) PostProcess(p schema.Pipeline[V]) bool {
	if successitems, ok := p.PostProcess(q.success); ok {
		q.success = successitems

		return true
	}

	return false
}
