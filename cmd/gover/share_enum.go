package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/desertwitch/gover/internal/queue"
	"github.com/desertwitch/gover/internal/schema"
)

// Enumerate is the principal method for querying [schema.Share] for candidate
// [schema.Moveable] and enqueueing them into a [queue.EvaluationManager]. This
// process happens concurrently, meaning multiple [schema.Storage] are read for
// a [schema.Share] (and its candidate [schema.Moveable]) at the same time.
func (app *app) Enumerate(ctx context.Context) error {
	tasker := queue.NewTaskManager()

	// Primary to Secondary
	for _, share := range app.shares {
		if share.GetUseCache() != "yes" || share.GetCachePool() == nil {
			continue
		}

		if share.GetCachePool2() == nil {
			// Cache to Array
			app.queueManager.EnumerationManager.Enqueue(&queue.EnumerationTask{
				Share:  share,
				Source: share.GetCachePool(),
				Function: func(share schema.Share, src schema.Storage, dst schema.Storage) func() int {
					return func() int {
						return app.enumerateToEvaluation(ctx, share, src, dst)
					}
				}(share, share.GetCachePool(), nil),
			})
		} else {
			// Cache to Cache2
			app.queueManager.EnumerationManager.Enqueue(&queue.EnumerationTask{
				Share:  share,
				Source: share.GetCachePool(),
				Function: func(share schema.Share, src schema.Storage, dst schema.Storage) func() int {
					return func() int {
						return app.enumerateToEvaluation(ctx, share, src, dst)
					}
				}(share, share.GetCachePool(), share.GetCachePool2()),
			})
		}
	}

	// Secondary to Primary
	for _, share := range app.shares {
		if share.GetUseCache() != "prefer" || share.GetCachePool() == nil {
			continue
		}

		if share.GetCachePool2() == nil {
			// Array to Cache
			for _, disk := range share.GetIncludedDisks() {
				app.queueManager.EnumerationManager.Enqueue(&queue.EnumerationTask{
					Share:  share,
					Source: disk,
					Function: func(share schema.Share, src schema.Storage, dst schema.Storage) func() int {
						return func() int {
							return app.enumerateToEvaluation(ctx, share, src, dst)
						}
					}(share, disk, share.GetCachePool()),
				})
			}
		} else {
			// Cache2 to Cache
			app.queueManager.EnumerationManager.Enqueue(&queue.EnumerationTask{
				Share:  share,
				Source: share.GetCachePool2(),
				Function: func(share schema.Share, src schema.Storage, dst schema.Storage) func() int {
					return func() int {
						return app.enumerateToEvaluation(ctx, share, src, dst)
					}
				}(share, share.GetCachePool2(), share.GetCachePool()),
			})
		}
	}

	for source, sourceQueue := range app.queueManager.EnumerationManager.GetQueues() {
		tasker.Add(func(source schema.Storage, sourceQueue *queue.EnumerationSourceQueue) func() {
			return func() {
				_ = app.processEnumerationQueue(ctx, source, sourceQueue)
			}
		}(source, sourceQueue))
	}

	if err := tasker.LaunchConcAndWait(ctx, runtime.NumCPU()); err != nil {
		return fmt.Errorf("(app-enum) %w", err)
	}

	return nil
}

// processEnumerationQueue processes an [queue.EnumerationSourceQueue]'s items
// (which are [queue.EnumerationTask] of one specific source [schema.Storage])
// and runs their contained enumeration functions concurrently. This means
// multiple [schema.Share] on one source [schema.Storage] are read for their
// [schema.Moveable] at the same time.
func (app *app) processEnumerationQueue(ctx context.Context, source schema.Storage, sourceQueue *queue.EnumerationSourceQueue) bool {
	slog.Info("Enumerating shares on source:",
		"source", source.GetName(),
	)

	if err := sourceQueue.DequeueAndProcessConc(ctx, runtime.NumCPU(), func(enumTask *queue.EnumerationTask) int {
		return enumTask.Run()
	}); err != nil {
		slog.Warn("Skipped enumerating shares on source due to failure:",
			"err", err,
			"source", source.GetName(),
		)

		return false
	}

	slog.Info("Enumerating shares on source done:",
		"share", source.GetName(),
	)

	return true
}

// enumerateToEvaluation is the actual given to [queue.EnumerationTask] function
// that collects all [schema.Moveable] for a [schema.Share] on a specific source
// [schema.Storage] and enqueues the results into the [queue.EvaluationManager].
func (app *app) enumerateToEvaluation(ctx context.Context, share schema.Share, src schema.Storage, dst schema.Storage) int {
	slog.Info("Enumerating share on storage:",
		"storage", src.GetName(),
		"share", share.GetName(),
	)

	files, err := app.fsHandler.GetMoveables(ctx, share, src, dst)
	if err != nil {
		slog.Warn("Skipped enumerating share on storage due to failure:",
			"err", err,
			"storage", src.GetName(),
			"share", share.GetName(),
		)

		return queue.DecisionSkipped
	}

	slog.Info("Enumerating shares on storage done:",
		"storage", src.GetName(),
		"share", share.GetName(),
	)

	app.queueManager.EvaluationManager.Enqueue(files...)

	return queue.DecisionSuccess
}
