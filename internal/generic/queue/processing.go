package queue

import (
	"context"
	"sync"
)

const (
	DecisionRequeue = -1
	DecisionSkipped = 0
	DecisionSuccess = 1
)

type queueProvider[T any] interface {
	Dequeue() (T, bool)
	Enqueue(items ...T)
	GetSuccessful() []T
	HasRemainingItems() bool
	SetProcessing(items ...T)
	SetSkipped(items ...T)
	SetSuccess(items ...T)
}

func processQueue[T any](ctx context.Context, queue queueProvider[T], processFunc func(T) int) error {
	for {
		if ctx.Err() != nil {
			break
		}

		item, ok := queue.Dequeue()
		if !ok {
			break
		}

		queue.SetProcessing(item)

		switch processFunc(item) {
		case DecisionRequeue:
			queue.Enqueue(item)

		case DecisionSkipped:
			queue.SetSkipped(item)

		case DecisionSuccess:
			queue.SetSuccess(item)
		}
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	return nil
}

func concurrentProcessQueue[T any](ctx context.Context, maxWorkers int, queue queueProvider[T], processFunc func(T) int) error {
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, maxWorkers)

LOOP:
	for {
		select {
		case <-ctx.Done():
			wg.Wait()

			return ctx.Err()
		case semaphore <- struct{}{}:
		}

		item, ok := queue.Dequeue()
		if !ok {
			<-semaphore

			break
		}

		wg.Add(1)
		go func(item T) {
			defer wg.Done()
			defer func() { <-semaphore }()

			queue.SetProcessing(item)

			switch processFunc(item) {
			case DecisionRequeue:
				queue.Enqueue(item)

			case DecisionSkipped:
				queue.SetSkipped(item)

			case DecisionSuccess:
				queue.SetSuccess(item)
			}
		}(item)
	}

	wg.Wait()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if queue.HasRemainingItems() {
		// In case item(s) were requeued but all workers have already left.
		goto LOOP
	}

	return nil
}
