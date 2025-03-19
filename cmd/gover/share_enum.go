package main

import (
	"context"
	"log/slog"
	"runtime"

	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/schema"
)

func (app *App) Enumerate(ctx context.Context) error {
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
				Function: func(share schema.Share, src schema.Storage, dst schema.Storage) func() error {
					return func() error {
						return app.enumerateToEvaluation(ctx, share, src, dst)
					}
				}(share, share.GetCachePool(), nil),
			})
		} else {
			// Cache to Cache2
			app.queueManager.EnumerationManager.Enqueue(&queue.EnumerationTask{
				Share:  share,
				Source: share.GetCachePool(),
				Function: func(share schema.Share, src schema.Storage, dst schema.Storage) func() error {
					return func() error {
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
					Function: func(share schema.Share, src schema.Storage, dst schema.Storage) func() error {
						return func() error {
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
				Function: func(share schema.Share, src schema.Storage, dst schema.Storage) func() error {
					return func() error {
						return app.enumerateToEvaluation(ctx, share, src, dst)
					}
				}(share, share.GetCachePool2(), share.GetCachePool()),
			})
		}
	}

	for shareName, shareQueue := range app.queueManager.EnumerationManager.GetQueues() {
		tasker.Add(func(shareName string, shareQueue *queue.EnumerationShareQueue) func() {
			return func() {
				_ = app.processEnumerationQueue(ctx, shareName, shareQueue)
			}
		}(shareName, shareQueue))
	}

	if err := tasker.LaunchConcAndWait(ctx, runtime.NumCPU()); err != nil {
		return err
	}

	return nil
}

func (app *App) processEnumerationQueue(ctx context.Context, shareName string, shareQueue *queue.EnumerationShareQueue) error {
	slog.Info("Enumerating share:",
		"share", shareName,
	)

	if err := shareQueue.DequeueAndProcessConc(ctx, runtime.NumCPU(), func(enumFunc *queue.EnumerationTask) int {
		if err := enumFunc.Run(); err != nil {
			return queue.DecisionSkipped
		}

		return queue.DecisionSuccess
	}); err != nil {
		slog.Warn("Skipped enumerating share due to failure:",
			"err", err,
			"share", shareName,
		)

		return err
	}

	slog.Info("Enumerating share done:",
		"share", shareName,
	)

	return nil
}

func (app *App) enumerateToEvaluation(ctx context.Context, share schema.Share, src schema.Storage, dst schema.Storage) error {
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

		return err
	}

	slog.Info("Enumerating share on storage done:",
		"storage", src.GetName(),
		"share", share.GetName(),
	)

	app.queueManager.EvaluationManager.Enqueue(files...)

	return nil
}
