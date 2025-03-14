package queue

import (
	"context"
	"sync"
)

type queueProvider[T any] interface {
	Dequeue() (T, bool)
	SetSuccess(items ...T)
	SetSkipped(items ...T)
	SetProcessing(items ...T)
	ResetQueue()
}

func Process[T any](ctx context.Context, queue queueProvider[T], processFunc func(T) bool, resetQueueAfter bool) error {
	if resetQueueAfter {
		defer queue.ResetQueue()
	}

	for {
		if ctx.Err() != nil {
			break
		}

		item, ok := queue.Dequeue()
		if !ok {
			break
		}

		queue.SetProcessing(item)

		if success := processFunc(item); success {
			queue.SetSuccess(item)
		} else {
			queue.SetSkipped(item)
		}
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	return nil
}

func ConcurrentProcess[T any](ctx context.Context, maxWorkers int, queue queueProvider[T], processFunc func(T) bool, resetQueueAfter bool) error {
	if resetQueueAfter {
		defer queue.ResetQueue()
	}

	var wg sync.WaitGroup

	semaphore := make(chan struct{}, maxWorkers)

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

			if success := processFunc(item); success {
				queue.SetSuccess(item)
			} else {
				queue.SetSkipped(item)
			}
		}(item)
	}

	wg.Wait()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	return nil
}
