package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/desertwitch/gover/internal/queue"
)

// IO is the principal method for moving all [schema.Moveable] to their
// respective target [schema.Storage]. This process happens concurrently,
// meaning multiple (different) [schema.Storage] get written to at the same
// time, but with only one I/O write operation ever happening per individual
// [schema.Storage] (= sequential processing inside one [schema.Storage]).
func (app *app) IO(ctx context.Context) error {
	tasker := queue.NewTaskManager()

	queues := app.queueManager.IOManager.GetQueues()

	for target, targetQueue := range queues {
		tasker.Add(
			func(targetQueue *queue.IOTargetQueue) func() {
				return func() {
					_ = app.ioHandler.ProcessTargetQueue(ctx, app.config.Pipelines.IOPipelines, target, targetQueue)
				}
			}(targetQueue),
		)
	}

	if err := tasker.LaunchConcAndWait(ctx, runtime.NumCPU()); err != nil {
		return fmt.Errorf("(app-io) %w", err)
	}

	return nil
}
