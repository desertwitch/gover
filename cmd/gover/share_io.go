package main

import (
	"context"
	"runtime"

	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/schema"
)

func (app *App) ioProcessFiles(ctx context.Context, files []*schema.Moveable) error {
	tasker := queue.NewTaskManager()

	app.queueManager.IOManager.Enqueue(files...)
	queues := app.queueManager.IOManager.GetQueuesUnsafe()

	for _, q := range queues {
		tasker.Add(
			func(q *queue.IOTargetQueue) func() {
				return func() {
					app.ioHandler.ProcessQueue(ctx, q)
				}
			}(q),
		)
	}

	if err := tasker.LaunchConcAndWait(ctx, runtime.NumCPU()); err != nil {
		return err
	}

	return nil
}
