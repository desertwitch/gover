package queue

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	DecisionRequeue = -1
	DecisionSkipped = 0
	DecisionSuccess = 1
)

type GenericQueue[T comparable] struct {
	sync.RWMutex
	hasStarted  bool
	hasFinished bool
	startTime   time.Time
	finishTime  time.Time
	head        int
	items       []T
	success     []T
	skipped     []T
	inProgress  map[T]struct{}
}

func NewGenericQueue[T comparable]() *GenericQueue[T] {
	return &GenericQueue[T]{
		inProgress: make(map[T]struct{}),
	}
}

func (q *GenericQueue[T]) HasRemainingItems() bool {
	q.RLock()
	defer q.RUnlock()

	if q.head >= len(q.items) {
		return false
	}

	return true
}

func (q *GenericQueue[T]) GetSuccessful() []T {
	q.RLock()
	defer q.RUnlock()

	result := make([]T, len(q.success))
	copy(result, q.success)

	return result
}

func (q *GenericQueue[T]) Enqueue(items ...T) {
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

func (q *GenericQueue[T]) Dequeue() (T, bool) { //nolint:ireturn
	q.Lock()
	defer q.Unlock()

	if q.head >= len(q.items) {
		var zeroVal T

		if !q.hasFinished {
			q.finishTime = time.Now()
			q.hasFinished = true
		}

		return zeroVal, false
	}

	if !q.hasStarted {
		q.startTime = time.Now()
		q.hasStarted = true
	}

	item := q.items[q.head]
	q.head++

	return item, true
}

func (q *GenericQueue[T]) SetSuccess(items ...T) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.success = append(q.success, item)
	}
}

func (q *GenericQueue[T]) SetSkipped(items ...T) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		delete(q.inProgress, item)
		q.skipped = append(q.skipped, item)
	}
}

func (q *GenericQueue[T]) SetProcessing(items ...T) {
	q.Lock()
	defer q.Unlock()

	for _, item := range items {
		q.inProgress[item] = struct{}{}
	}
}

func (q *GenericQueue[T]) Progress() Progress {
	q.RLock()
	defer q.RUnlock()

	hasStarted := q.hasStarted
	totalItems := len(q.items)

	processedItems := q.head
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

func (q *GenericQueue[T]) DequeueAndProcess(ctx context.Context, processFunc func(T) int) error {
	for {
		if ctx.Err() != nil {
			break
		}

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

func (q *GenericQueue[T]) DequeueAndProcessConc(ctx context.Context, maxWorkers int, processFunc func(T) int) error {
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
		go func(item T) {
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
