package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/schema"
)

func (app *App) Enumerate(ctx context.Context) error { //nolint:funlen
	tasker := queue.NewTaskManager()

	enumKeyFunc := func(et *queue.EnumerationTask) string { return et.Source.GetName() }
	enumFactoryFunc := queue.NewEnumerationSourceQueue

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
			}, enumKeyFunc, enumFactoryFunc)
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
			}, enumKeyFunc, enumFactoryFunc)
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
				}, enumKeyFunc, enumFactoryFunc)
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
			}, enumKeyFunc, enumFactoryFunc)
		}
	}

	for sourceName, sourceQueue := range app.queueManager.EnumerationManager.GetQueues() {
		tasker.Add(func(sourceName string, sourceQueue *queue.EnumerationSourceQueue) func() {
			return func() {
				_ = app.processEnumerationQueue(ctx, sourceName, sourceQueue)
			}
		}(sourceName, sourceQueue))
	}

	if err := tasker.LaunchConcAndWait(ctx, runtime.NumCPU()); err != nil {
		return fmt.Errorf("(app-enum) %w", err)
	}

	return nil
}

func (app *App) processEnumerationQueue(ctx context.Context, sourceName string, sourceQueue *queue.EnumerationSourceQueue) bool {
	slog.Info("Enumerating shares on source:",
		"source", sourceName,
	)

	if err := sourceQueue.DequeueAndProcessConc(ctx, runtime.NumCPU(), func(enumTask *queue.EnumerationTask) int {
		return enumTask.Run()
	}); err != nil {
		slog.Warn("Skipped enumerating shares on source due to failure:",
			"err", err,
			"source", sourceName,
		)

		return false
	}

	slog.Info("Enumerating shares on source done:",
		"share", sourceName,
	)

	return true
}

func (app *App) enumerateToEvaluation(ctx context.Context, share schema.Share, src schema.Storage, dst schema.Storage) int {
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

	app.queueManager.EvaluationManager.EnqueueMany(files, func(m *schema.Moveable) string {
		return m.Share.GetName()
	}, queue.NewEvaluationShareQueue)

	return queue.DecisionSuccess
}
