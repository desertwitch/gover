package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/desertwitch/gover/internal/generic/queue"
)

func (app *App) IO(ctx context.Context) error {
	tasker := queue.NewTaskManager()

	queues := app.queueManager.IOManager.GetQueues()

	for _, targetQueue := range queues {
		tasker.Add(
			func(targetQueue *queue.IOTargetQueue) func() {
				return func() {
					_ = app.ioHandler.ProcessTargetQueue(ctx, targetQueue)
				}
			}(targetQueue),
		)
	}

	if err := tasker.LaunchConcAndWait(ctx, runtime.NumCPU()); err != nil {
		return fmt.Errorf("(app-io) %w", err)
	}

	return nil
}
