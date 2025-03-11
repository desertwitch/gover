package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
	"github.com/desertwitch/gover/internal/validation"
)

func enumSystem(ctx context.Context, system *unraid.System, deps *depCoordinator) ([]*filesystem.Moveable, error) {
	var wg sync.WaitGroup

	tasks := []func(){}
	ch := make(chan []*filesystem.Moveable, len(system.Shares))

	// Primary to Secondary
	for _, share := range system.Shares {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if share.UseCache != "yes" || share.CachePool == nil {
			continue
		}

		if share.CachePool2 == nil {
			// Cache to Array
			tasks = append(tasks, func() {
				newShareEnumWorker(ch, share, share.CachePool, nil, deps)
			})
		} else {
			// Cache to Cache2
			tasks = append(tasks, func() {
				newShareEnumWorker(ch, share, share.CachePool, share.CachePool2, deps)
			})
		}
	}

	// Secondary to Primary
	for _, share := range system.Shares {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if share.UseCache != "prefer" || share.CachePool == nil {
			continue
		}

		if share.CachePool2 == nil {
			// Array to Cache
			for _, disk := range system.Array.Disks {
				tasks = append(tasks, func() {
					newShareEnumWorker(ch, share, disk, share.CachePool, deps)
				})
			}
		} else {
			// Cache2 to Cache
			tasks = append(tasks, func() {
				newShareEnumWorker(ch, share, share.CachePool2, share.CachePool, deps)
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

func newShareEnumWorker(ch chan<- []*filesystem.Moveable, share *unraid.Share, src unraid.Storeable, dst unraid.Storeable, deps *depCoordinator) {
	files, err := enumShare(share, src, dst, deps)
	if err != nil {
		if _, ok := src.(*unraid.Disk); ok {
			slog.Warn("Skipped processing array disk due to failure",
				"err", err,
				"share", share.Name,
			)
		} else {
			slog.Warn("Skipped processing share due to failure",
				"err", err,
				"share", share.Name,
			)
		}

		return
	}

	ch <- files
}

func enumShare(share *unraid.Share, src unraid.Storeable, dst unraid.Storeable, deps *depCoordinator) ([]*filesystem.Moveable, error) {
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
