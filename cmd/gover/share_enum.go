package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"

	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/schema"
	"github.com/desertwitch/gover/internal/generic/validation"
)

func enumerateShares(ctx context.Context, shares map[string]schema.Share, queueMan *queue.Manager, deps *depPackage) ([]*filesystem.Moveable, error) {
	var wg sync.WaitGroup

	tasks := []func(){}

	// Primary to Secondary
	for _, share := range shares {
		if share.GetUseCache() != "yes" || share.GetCachePool() == nil {
			continue
		}

		if share.GetCachePool2() == nil {
			// Cache to Array
			tasks = append(tasks, func() {
				shareEnumerationWorker(ctx, share, share.GetCachePool(), nil, queueMan, deps)
			})
		} else {
			// Cache to Cache2
			tasks = append(tasks, func() {
				shareEnumerationWorker(ctx, share, share.GetCachePool(), share.GetCachePool2(), queueMan, deps)
			})
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
				tasks = append(tasks, func() {
					shareEnumerationWorker(ctx, share, disk, share.GetCachePool(), queueMan, deps)
				})
			}
		} else {
			// Cache2 to Cache
			tasks = append(tasks, func() {
				shareEnumerationWorker(ctx, share, share.GetCachePool2(), share.GetCachePool(), queueMan, deps)
			})
		}
	}

	maxWorkers := runtime.NumCPU()
	semaphore := make(chan struct{}, maxWorkers)

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			wg.Wait()

			return nil, ctx.Err()
		case semaphore <- struct{}{}:
		}

		wg.Add(1)
		go func(task func()) {
			defer wg.Done()
			defer func() { <-semaphore }()

			task()
		}(task)
	}

	wg.Wait()

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
	files, err := deps.FSHandler.GetMoveables(ctx, share, src, dst)
	if err != nil {
		return fmt.Errorf("(main) failed to enumerate: %w", err)
	}

	q := queueMan.EnumerationManager.NewQueue()
	q.Enqueue(files...)

	if dst == nil {
		if err = deps.AllocHandler.AllocateArrayDestinations(ctx, q); err != nil {
			return fmt.Errorf("(main) failed to allocate: %w", err)
		}
	}

	if err := deps.PathingHandler.EstablishPaths(ctx, q); err != nil {
		return fmt.Errorf("(main) failed to establish paths: %w", err)
	}

	if err := validation.ValidateMoveables(ctx, q); err != nil {
		return fmt.Errorf("(main) failed to validate: %w", err)
	}

	return nil
}
