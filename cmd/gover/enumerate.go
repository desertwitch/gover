package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
	"github.com/desertwitch/gover/internal/validation"
)

func getFilesBySystem(ctx context.Context, system *unraid.System, handlers *taskHandlers) ([]*filesystem.Moveable, error) {
	var wg sync.WaitGroup

	filechan := make(chan []*filesystem.Moveable, len(system.Shares))

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
			wg.Add(1)
			go func(share *unraid.Share) {
				defer wg.Done()

				sfiles, err := getFilesByShare(share, share.CachePool, nil, handlers)
				if err != nil {
					slog.Warn("Skipped processing share due to failure",
						"err", err,
						"share", share.Name,
					)

					return
				}

				filechan <- sfiles
			}(share)
		} else {
			// Cache to Cache2

			wg.Add(1)
			go func(share *unraid.Share) {
				defer wg.Done()

				sfiles, err := getFilesByShare(share, share.CachePool, share.CachePool2, handlers)
				if err != nil {
					slog.Warn("Skipped processing share due to failure",
						"err", err,
						"share", share.Name,
					)

					return
				}

				filechan <- sfiles
			}(share)
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
				wg.Add(1)
				go func(share *unraid.Share) {
					defer wg.Done()

					sfiles, err := getFilesByShare(share, disk, share.CachePool, handlers)
					if err != nil {
						slog.Warn("Skipped processing array disk of share due to failure",
							"err", err,
							"share", share.Name,
						)

						return
					}

					filechan <- sfiles
				}(share)
			}
		} else {
			// Cache2 to Cache
			wg.Add(1)
			go func(share *unraid.Share) {
				defer wg.Done()

				sfiles, err := getFilesByShare(share, share.CachePool2, share.CachePool, handlers)
				if err != nil {
					slog.Warn("Skipped processing share due to failure",
						"err", err,
						"share", share.Name,
					)

					return
				}

				filechan <- sfiles
			}(share)
		}
	}

	go func() {
		wg.Wait()
		close(filechan)
	}()

	var files []*filesystem.Moveable
	for sharefiles := range filechan {
		files = append(files, sharefiles...)
	}

	return files, nil
}

func getFilesByShare(share *unraid.Share, src unraid.Storeable, dst unraid.Storeable, deps *taskHandlers) ([]*filesystem.Moveable, error) {
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
