package main

import (
	"context"
	"log/slog"
	"runtime"

	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/schema"
	"github.com/desertwitch/gover/internal/generic/validation"
)

func (app *App) Evaluate(ctx context.Context) error {
	tasker := queue.NewTaskManager()

	queues := app.queueManager.EvaluationManager.GetQueues()

	for name, q := range queues {
		tasker.Add(
			func(name string, q *queue.EvaluationShareQueue) func() {
				return func() {
					slog.Info("Evaluating share:",
						"share", name,
					)

					if err := app.evaluateToIO(ctx, q); err != nil {
						slog.Warn("Skipped evaluating share due to failure:",
							"err", err,
							"share", name,
						)

						return
					}

					slog.Info("Evaluating share done:",
						"share", name,
					)
				}
			}(name, q),
		)
	}

	if err := tasker.LaunchConcAndWait(ctx, runtime.NumCPU()); err != nil {
		return err
	}

	return nil
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
		return err
	}

	app.queueManager.IOManager.Enqueue(q.GetSuccessful()...)

	return nil
}
