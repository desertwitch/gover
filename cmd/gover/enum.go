package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/validation"
)

func enumerateShares(ctx context.Context, shares map[string]filesystem.ShareType, disks map[string]filesystem.DiskType, deps *depPackage) ([]*filesystem.Moveable, error) {
	var wg sync.WaitGroup

	tasks := []func(){}
	ch := make(chan []*filesystem.Moveable, len(shares))

	// Primary to Secondary
	for _, share := range shares {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if share.GetUseCache() != "yes" || share.GetCachePool() == nil {
			continue
		}

		if share.GetCachePool2() == nil {
			// Cache to Array
			tasks = append(tasks, func() {
				enumerationWorker(ch, share, share.GetCachePool(), nil, deps)
			})
		} else {
			// Cache to Cache2
			tasks = append(tasks, func() {
				enumerationWorker(ch, share, share.GetCachePool(), share.GetCachePool2(), deps)
			})
		}
	}

	// Secondary to Primary
	for _, share := range shares {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if share.GetUseCache() != "prefer" || share.GetCachePool() == nil {
			continue
		}

		if share.GetCachePool2() == nil {
			// Array to Cache
			for _, disk := range disks {
				tasks = append(tasks, func() {
					enumerationWorker(ch, share, disk, share.GetCachePool(), deps)
				})
			}
		} else {
			// Cache2 to Cache
			tasks = append(tasks, func() {
				enumerationWorker(ch, share, share.GetCachePool2(), share.GetCachePool(), deps)
			})
		}
	}

	maxWorkers := runtime.NumCPU()
	semaphore := make(chan struct{}, maxWorkers)

	for _, task := range tasks {
		semaphore <- struct{}{}

		wg.Add(1)
		go func(task func()) {
			defer wg.Done()
			defer func() { <-semaphore }()
			task()
		}(task)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var files []*filesystem.Moveable
	for f := range ch {
		files = append(files, f...)
	}

	return files, nil
}

func enumerationWorker(ch chan<- []*filesystem.Moveable, share filesystem.ShareType, src filesystem.StorageType, dst filesystem.StorageType, deps *depPackage) {
	files, err := enumerateShare(share, src, dst, deps)
	if err != nil {
		if _, ok := src.(filesystem.DiskType); ok {
			slog.Warn("Skipped processing array disk due to failure",
				"err", err,
				"share", share.GetName(),
			)
		} else {
			slog.Warn("Skipped processing share due to failure",
				"err", err,
				"share", share.GetName(),
			)
		}

		return
	}

	ch <- files
}

func enumerateShare(share filesystem.ShareType, src filesystem.StorageType, dst filesystem.StorageType, deps *depPackage) ([]*filesystem.Moveable, error) {
	files, err := deps.FSHandler.GetMoveables(share, src, dst)
	if err != nil {
		return nil, fmt.Errorf("(main) failed to enumerate: %w", err)
	}

	if dst == nil {
		files, err = deps.AllocHandler.AllocateArrayDestinations(files)
		if err != nil {
			return nil, fmt.Errorf("(main) failed to allocate: %w", err)
		}
	}

	files, err = deps.FSHandler.EstablishPaths(files)
	if err != nil {
		return nil, fmt.Errorf("(main) failed to establish paths: %w", err)
	}

	files, err = validation.ValidateMoveables(files)
	if err != nil {
		return nil, fmt.Errorf("(main) failed to validate: %w", err)
	}

	return files, nil
}
