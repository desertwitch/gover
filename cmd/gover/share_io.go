package main

import (
	"context"
	"runtime"
	"sync"

	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/schema"
)

func ioProcessFiles(ctx context.Context, files []*schema.Moveable, queueMan *queue.Manager, deps *depPackage) error {
	queueMan.IOManager.Enqueue(files...)

	queues := queueMan.IOManager.GetQueuesUnsafe()

	var queueWG sync.WaitGroup

	maxWorkers := runtime.NumCPU()
	semaphore := make(chan struct{}, maxWorkers)

	for _, q := range queues {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case semaphore <- struct{}{}:
		}

		queueWG.Add(1)
		go func(q *queue.IOTargetQueue) {
			defer queueWG.Done()
			defer func() { <-semaphore }()

			deps.IOHandler.ProcessQueue(ctx, q)
		}(q)
	}

	queueWG.Wait()

	return nil
}
