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

func (app *App) enumerateShares(ctx context.Context, shares map[string]schema.Share) ([]*schema.Moveable, error) {
	tasker := queue.NewTaskManager()

	// Primary to Secondary
	for _, share := range shares {
		if share.GetUseCache() != "yes" || share.GetCachePool() == nil {
			continue
		}

		if share.GetCachePool2() == nil {
			// Cache to Array
			tasker.Add(
				func(share schema.Share, src schema.Storage, dst schema.Storage) func() {
					return func() {
						_ = app.enumerateShare(ctx, share, src, dst)
					}
				}(share, share.GetCachePool(), nil),
			)
		} else {
			// Cache to Cache2
			tasker.Add(
				func(share schema.Share, src schema.Storage, dst schema.Storage) func() {
					return func() {
						_ = app.enumerateShare(ctx, share, src, dst)
					}
				}(share, share.GetCachePool(), share.GetCachePool2()),
			)
		}
	}

	// Secondary to Primary
	for _, share := range shares {
		if share.GetUseCache() != "prefer" || share.GetCachePool() == nil {
			continue
		}

		if share.GetCachePool2() == nil {
			// Array to Cache
			for _, disk := range share.GetIncludedDisks() {
				tasker.Add(
					func(share schema.Share, src schema.Storage, dst schema.Storage) func() {
						return func() {
							_ = app.enumerateShare(ctx, share, src, dst)
						}
					}(share, disk, share.GetCachePool()),
				)
			}
		} else {
			// Cache2 to Cache
			tasker.Add(
				func(share schema.Share, src schema.Storage, dst schema.Storage) func() {
					return func() {
						_ = app.enumerateShare(ctx, share, src, dst)
					}
				}(share, share.GetCachePool2(), share.GetCachePool()),
			)
		}
	}

	if err := tasker.LaunchConcAndWait(ctx, runtime.NumCPU()); err != nil {
		return nil, err
	}

	return app.queueManager.EnumerationManager.GetItems(), nil
}

func (app *App) enumerateShare(ctx context.Context, share schema.Share, src schema.Storage, dst schema.Storage) error {
	slog.Info("Enumerating share on storage:", "src", src.GetName(), "share", share.GetName())

	if err := app.enumerateShareTask(ctx, share, src, dst); err != nil {
		if _, ok := src.(schema.Disk); ok {
			slog.Warn("Skipped enumerating array disk due to failure",
				"err", err,
				"share", share.GetName(),
			)
		} else {
			slog.Warn("Skipped enumerating share due to failure",
				"err", err,
				"share", share.GetName(),
			)
		}

		return err
	}

	slog.Info("Enumerating share on storage done:", "src", src.GetName(), "share", share.GetName())

	return nil
}

func (app *App) enumerateShareTask(ctx context.Context, share schema.Share, src schema.Storage, dst schema.Storage) error {
	q := app.queueManager.EnumerationManager.NewQueue()

	app.queueManager.EnumerationManager.SetQueuePhase(q, "initializing")

	files, err := app.fsHandler.GetMoveables(ctx, share, src, dst)
	if err != nil {
		app.queueManager.EnumerationManager.DestroyQueue(q)

		return fmt.Errorf("(main) failed to enumerate: %w", err)
	}

	q.Enqueue(files...)

	if dst == nil {
		app.queueManager.EnumerationManager.SetQueuePhase(q, "allocating")
		if err = app.allocHandler.AllocateArrayDestinations(ctx, q); err != nil {
			app.queueManager.EnumerationManager.DestroyQueue(q)

			return fmt.Errorf("(main) failed to allocate: %w", err)
		}
	}

	app.queueManager.EnumerationManager.SetQueuePhase(q, "pathing")
	if err := app.pathingHandler.EstablishPaths(ctx, q); err != nil {
		app.queueManager.EnumerationManager.DestroyQueue(q)

		return fmt.Errorf("(main) failed to establish paths: %w", err)
	}

	app.queueManager.EnumerationManager.SetQueuePhase(q, "validating")
	if err := validation.ValidateMoveables(ctx, q); err != nil {
		app.queueManager.EnumerationManager.DestroyQueue(q)

		return fmt.Errorf("(main) failed to validate: %w", err)
	}

	app.queueManager.EnumerationManager.SetQueuePhase(q, "waiting for IO")

	return nil
}
