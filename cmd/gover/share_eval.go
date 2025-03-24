package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/schema"
	"github.com/desertwitch/gover/internal/generic/validation"
)

func (app *App) Evaluate(ctx context.Context) error {
	tasker := queue.NewTaskManager()

	for shareName, shareQueue := range app.queueManager.EvaluationManager.GetQueues() {
		tasker.Add(
			func(shareName string, shareQueue *queue.EvaluationShareQueue) func() {
				return func() {
					_ = app.processEvaluationQueue(ctx, shareName, shareQueue)
				}
			}(shareName, shareQueue),
		)
	}

	if err := tasker.LaunchConcAndWait(ctx, runtime.NumCPU()); err != nil {
		return fmt.Errorf("(app-eval) %w", err)
	}

	return nil
}

func (app *App) processEvaluationQueue(ctx context.Context, shareName string, shareQueue *queue.EvaluationShareQueue) bool {
	slog.Info("Evaluating share:",
		"share", shareName,
	)

	if err := app.evaluateToIO(ctx, shareQueue); err != nil {
		slog.Warn("Skipped evaluating share due to failure:",
			"err", err,
			"share", shareName,
		)

		return false
	}

	slog.Info("Evaluating share done:",
		"share", shareName,
	)

	return true
}

func (app *App) evaluateToIO(ctx context.Context, q *queue.EvaluationShareQueue) error {
	if err := q.DequeueAndProcessConc(ctx, runtime.NumCPU(), func(m *schema.Moveable) int {
		if m.Dest == nil {
			if success := app.allocHandler.AllocateArrayDestination(m); !success {
				return queue.DecisionSkipped
			}
		}

		if success := app.pathingHandler.EstablishPath(m); !success {
			return queue.DecisionSkipped
		}

		if success := validation.ValidateMoveable(m); !success {
			return queue.DecisionSkipped
		}

		return queue.DecisionSuccess
	}); err != nil {
		return fmt.Errorf("(app-eval) %w", err)
	}

	app.queueManager.IOManager.EnqueueMany(q.GetSuccessful(), func(m *schema.Moveable) string {
		return m.Dest.GetName()
	}, queue.NewIOTargetQueue)

	return nil
}
