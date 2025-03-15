package main

import (
	"context"
	"runtime"

	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/schema"
)

func ioProcessFiles(ctx context.Context, files []*schema.Moveable, queueMan *queue.Manager, deps *depPackage) error {
	tasker := queue.NewTaskManager()

	queueMan.IOManager.Enqueue(files...)
	queues := queueMan.IOManager.GetQueuesUnsafe()

	for _, q := range queues {
		tasker.Add(
			func(q *queue.IOTargetQueue) func() {
				return func() {
					deps.IOHandler.ProcessQueue(ctx, q)
				}
			}(q),
		)
	}

	if err := tasker.LaunchConcAndWait(ctx, runtime.NumCPU()); err != nil {
		return err
	}

	return nil
}
