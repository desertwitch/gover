package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"

	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/storage"
	"github.com/desertwitch/gover/internal/generic/validation"
)

func processShares(ctx context.Context, wg *sync.WaitGroup, shares map[string]storage.Share, deps *depPackage) {
	defer wg.Done()

	queueMan := queue.NewManager()
	files := enumerateShares(ctx, shares, deps)

	queueMan.Enqueue(files...)
	destQueues := queueMan.GetQueuesUnsafe()

	var queueWG sync.WaitGroup

	maxWorkers := runtime.NumCPU()
	semaphore := make(chan struct{}, maxWorkers)

	for _, destQueue := range destQueues {
		select {
		case <-ctx.Done():
			return
		case semaphore <- struct{}{}:
		}

		queueWG.Add(1)
		go func(q *queue.DestinationQueue) {
			defer queueWG.Done()
			defer func() { <-semaphore }()

			deps.IOHandler.ProcessQueue(ctx, q)
		}(destQueue)
	}

	queueWG.Wait()
}

//nolint:funlen
func enumerateShares(ctx context.Context, shares map[string]storage.Share, deps *depPackage) []*filesystem.Moveable {
	var wg sync.WaitGroup

	tasks := []func(){}
	ch := make(chan []*filesystem.Moveable)

	// Primary to Secondary
	for _, share := range shares {
		if share.GetUseCache() != "yes" || share.GetCachePool() == nil {
			continue
		}

		if share.GetCachePool2() == nil {
			// Cache to Array
			tasks = append(tasks, func() {
				shareEnumerationWorker(ctx, ch, share, share.GetCachePool(), nil, deps)
			})
		} else {
			// Cache to Cache2
			tasks = append(tasks, func() {
				shareEnumerationWorker(ctx, ch, share, share.GetCachePool(), share.GetCachePool2(), deps)
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
					shareEnumerationWorker(ctx, ch, share, disk, share.GetCachePool(), deps)
				})
			}
		} else {
			// Cache2 to Cache
			tasks = append(tasks, func() {
				shareEnumerationWorker(ctx, ch, share, share.GetCachePool2(), share.GetCachePool(), deps)
			})
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		maxWorkers := runtime.NumCPU()
		semaphore := make(chan struct{}, maxWorkers)

		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return
			case semaphore <- struct{}{}:
			}

			wg.Add(1)
			go func(task func()) {
				defer wg.Done()
				defer func() { <-semaphore }()

				task()
			}(task)
		}
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	files := []*filesystem.Moveable{}
	for f := range ch {
		files = append(files, f...)
	}

	return files
}

func shareEnumerationWorker(ctx context.Context, ch chan<- []*filesystem.Moveable, share storage.Share, src storage.Storage, dst storage.Storage, deps *depPackage) {
	slog.Info("Enumerating share on storage:", "src", src.GetName(), "share", share.GetName())

	files, err := enumerateShare(ctx, share, src, dst, deps)
	if err != nil {
		if _, ok := src.(storage.Disk); ok {
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

	select {
	case <-ctx.Done():
		return
	case ch <- files:
	}

	slog.Info("Enumerating share on storage done:", "src", src.GetName(), "share", share.GetName())
}

func enumerateShare(ctx context.Context, share storage.Share, src storage.Storage, dst storage.Storage, deps *depPackage) ([]*filesystem.Moveable, error) {
	files, err := deps.FSHandler.GetMoveables(ctx, share, src, dst)
	if err != nil {
		return nil, fmt.Errorf("(main) failed to enumerate: %w", err)
	}

	if dst == nil {
		files, err = deps.AllocHandler.AllocateArrayDestinations(ctx, files)
		if err != nil {
			return nil, fmt.Errorf("(main) failed to allocate: %w", err)
		}
	}

	files, err = deps.FSHandler.EstablishPaths(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("(main) failed to establish paths: %w", err)
	}

	files, err = validation.ValidateMoveables(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("(main) failed to validate: %w", err)
	}

	return files, nil
}
