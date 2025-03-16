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

func enumerateShares(ctx context.Context, shares map[string]schema.Share, queueMan *queue.Manager, deps *depPackage) ([]*schema.Moveable, error) {
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
						shareEnumerationWorker(ctx, share, src, dst, queueMan, deps)
					}
				}(share, share.GetCachePool(), nil),
			)
		} else {
			// Cache to Cache2
			tasker.Add(
				func(share schema.Share, src schema.Storage, dst schema.Storage) func() {
					return func() {
						shareEnumerationWorker(ctx, share, src, dst, queueMan, deps)
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
			for name, disk := range share.GetIncludedDisks() {
				if _, exists := share.GetExcludedDisks()[name]; exists {
					continue
				}
				tasker.Add(
					func(share schema.Share, src schema.Storage, dst schema.Storage) func() {
						return func() {
							shareEnumerationWorker(ctx, share, src, dst, queueMan, deps)
						}
					}(share, disk, share.GetCachePool()),
				)
			}
		} else {
			// Cache2 to Cache
			tasker.Add(
				func(share schema.Share, src schema.Storage, dst schema.Storage) func() {
					return func() {
						shareEnumerationWorker(ctx, share, src, dst, queueMan, deps)
					}
				}(share, share.GetCachePool2(), share.GetCachePool()),
			)
		}
	}

	if err := tasker.LaunchConcAndWait(ctx, runtime.NumCPU()); err != nil {
		return nil, err
	}

	return queueMan.EnumerationManager.GetItems(), nil
}

func shareEnumerationWorker(ctx context.Context, share schema.Share, src schema.Storage, dst schema.Storage, queueMan *queue.Manager, deps *depPackage) {
	slog.Info("Enumerating share on storage:", "src", src.GetName(), "share", share.GetName())

	if err := enumerateShare(ctx, share, src, dst, queueMan, deps); err != nil {
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

		return
	}

	slog.Info("Enumerating share on storage done:", "src", src.GetName(), "share", share.GetName())
}

func enumerateShare(ctx context.Context, share schema.Share, src schema.Storage, dst schema.Storage, queueMan *queue.Manager, deps *depPackage) error {
	q := queueMan.EnumerationManager.NewQueue()
	queueMan.EnumerationManager.SetQueuePhase(q, "initializing")

	files, err := deps.FSHandler.GetMoveables(ctx, share, src, dst)
	if err != nil {
		return fmt.Errorf("(main) failed to enumerate: %w", err)
	}

	q.Enqueue(files...)

	if dst == nil {
		queueMan.EnumerationManager.SetQueuePhase(q, "allocating")
		if err = deps.AllocHandler.AllocateArrayDestinations(ctx, q); err != nil {
			return fmt.Errorf("(main) failed to allocate: %w", err)
		}
	}

	queueMan.EnumerationManager.SetQueuePhase(q, "pathing")
	if err := deps.PathingHandler.EstablishPaths(ctx, q); err != nil {
		return fmt.Errorf("(main) failed to establish paths: %w", err)
	}

	queueMan.EnumerationManager.SetQueuePhase(q, "validating")
	if err := validation.ValidateMoveables(ctx, q); err != nil {
		return fmt.Errorf("(main) failed to validate: %w", err)
	}

	queueMan.EnumerationManager.SetQueuePhase(q, "waiting for IO")

	return nil
}
