package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/desertwitch/gover/internal/queue"
	"github.com/desertwitch/gover/internal/schema"
	"github.com/desertwitch/gover/internal/validation"
)

// Evaluate is the principal method for filtering, allocating, pathing and
// validating all previously collected [schema.Moveable] and enqueueing them
// into the [queue.IOManager]. This process happens concurrently, meaning that
// multiple [schema.Share]'s are processed at the same time.
func (app *app) Evaluate(ctx context.Context) error {
	tasker := queue.NewTaskManager()

	for share, shareQueue := range app.queueManager.EvaluationManager.GetQueues() {
		tasker.Add(
			func(share schema.Share, shareQueue *queue.EvaluationShareQueue) func() {
				return func() {
					_ = app.processEvaluationQueue(ctx, share, shareQueue)
				}
			}(share, shareQueue),
		)
	}

	if err := tasker.LaunchConcAndWait(ctx, runtime.NumCPU()); err != nil {
		return fmt.Errorf("(app-eval) %w", err)
	}

	return nil
}

// processEvaluationQueue processes an [queue.EvaluationShareQueue]'s items
// (which are [schema.Moveable] of one specific [schema.Share]).
func (app *app) processEvaluationQueue(ctx context.Context, share schema.Share, shareQueue *queue.EvaluationShareQueue) bool {
	slog.Info("Evaluating share:",
		"share", share.GetName(),
	)

	if err := app.evaluateToIO(ctx, share, shareQueue); err != nil {
		slog.Warn("Skipped evaluating share due to failure:",
			"err", err,
			"share", share.GetName(),
		)

		return false
	}

	slog.Info("Evaluating share done:",
		"share", share.GetName(),
	)

	return true
}

// evaluateToIO is the processing logic for an [queue.EvaluationShareQueue]. It
// processes items of the [queue.EvaluationShareQueue] concurrently, meaning
// that multiple [schema.Moveable] of one specific [schema.Share] are processed
// at the same time.
func (app *app) evaluateToIO(ctx context.Context, share schema.Share, q *queue.EvaluationShareQueue) error {
	if pipeline, exists := app.config.Pipelines.EvaluationPipelines[share.GetName()]; exists {
		if success := q.PreProcess(pipeline); !success {
			return fmt.Errorf("(app-eval) %w", ErrPipePreProcFailed)
		}
	}

	if err := q.DequeueAndProcessConc(ctx, runtime.NumCPU(), func(m *schema.Moveable) int {
		if pipeline, exists := app.config.Pipelines.EvaluationPipelines[share.GetName()]; exists {
			if success := pipeline.Process(m); !success {
				return queue.DecisionSkipped
			}
		}

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

	if pipeline, exists := app.config.Pipelines.EvaluationPipelines[share.GetName()]; exists {
		if success := q.PostProcess(pipeline); !success {
			return fmt.Errorf("(app-eval) %w", ErrPipePostProcFailed)
		}
	}

	app.queueManager.IOManager.Enqueue(q.GetSuccessful()...)

	return nil
}
