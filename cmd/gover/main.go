package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/desertwitch/gover/internal/allocation"
	"github.com/desertwitch/gover/internal/configuration"
	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/io"
	"github.com/desertwitch/gover/internal/unraid"
	"github.com/desertwitch/gover/internal/validation"
	"github.com/lmittmann/tint"
)

type taskHandlers struct {
	FSHandler     *filesystem.Handler
	UnraidHandler *unraid.Handler
	AllocHandler  *allocation.Handler
	IOHandler     *io.Handler
}

func moveShares(ctx context.Context, system *unraid.System, handlers *taskHandlers) {
	// Primary to Secondary
	for _, share := range system.Shares {
		if ctx.Err() != nil {
			return
		}

		if share.UseCache != "yes" || share.CachePool == nil {
			continue
		}

		if share.CachePool2 == nil {
			// Cache to Array
			if err := moveShare(ctx, share, share.CachePool, nil, handlers); err != nil {
				slog.Warn("Skipped share due to failure",
					"err", err,
					"share", share.Name,
				)

				continue
			}
		} else {
			// Cache to Cache2
			if err := moveShare(ctx, share, share.CachePool, share.CachePool2, handlers); err != nil {
				slog.Warn("Skipped share due to failure",
					"err", err,
					"share", share.Name,
				)

				continue
			}
		}
	}

	// Secondary to Primary
	for _, share := range system.Shares {
		if ctx.Err() != nil {
			return
		}

		if share.UseCache != "prefer" || share.CachePool == nil {
			continue
		}

		if share.CachePool2 == nil {
			// Array to Cache
			for _, disk := range system.Array.Disks {
				if err := moveShare(ctx, share, disk, share.CachePool, handlers); err != nil {
					slog.Warn("Skipped array disk of share due to failure",
						"err", err,
						"share", share.Name,
					)

					continue
				}
			}
		} else {
			// Cache2 to Cache
			if err := moveShare(ctx, share, share.CachePool2, share.CachePool, handlers); err != nil {
				slog.Warn("Skipped share due to failure",
					"err", err,
					"share", share.Name,
				)

				continue
			}
		}
	}
}

func moveShare(ctx context.Context, share *unraid.Share, src unraid.Storeable, dst unraid.Storeable, deps *taskHandlers) error {
	files, err := deps.FSHandler.GetMoveables(share, src, dst)
	if err != nil {
		return fmt.Errorf("failed to get moveables: %w", err)
	}

	if dst == nil {
		files, err = deps.AllocHandler.AllocateArrayDestinations(files)
		if err != nil {
			return fmt.Errorf("failed to allocate array destinations: %w", err)
		}
	}

	files, err = deps.FSHandler.EstablishPaths(files)
	if err != nil {
		return fmt.Errorf("failed to establish paths: %w", err)
	}

	files, err = validation.ValidateMoveables(files)
	if err != nil {
		return fmt.Errorf("failed to validate moveables: %w", err)
	}

	if err := deps.IOHandler.ProcessMoveables(ctx, files, &io.ProgressReport{}); err != nil {
		return fmt.Errorf("failed to move moveables: %w", err)
	}

	return nil
}

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		cancel()
	}()

	osProvider := &filesystem.OS{}
	unixProvider := &filesystem.Unix{}
	configProvider := &configuration.GodotenvProvider{}

	configOps := configuration.NewHandler(configProvider)
	fsOps := filesystem.NewHandler(osProvider, unixProvider)
	unraidOps := unraid.NewHandler(fsOps, configOps)
	allocOps := allocation.NewHandler(fsOps)
	ioOps := io.NewHandler(allocOps, fsOps, osProvider, unixProvider)

	deps := &taskHandlers{
		FSHandler:     fsOps,
		UnraidHandler: unraidOps,
		AllocHandler:  allocOps,
		IOHandler:     ioOps,
	}

	system, err := unraidOps.EstablishSystem()
	if err != nil {
		slog.Error("failed to establish unraid system",
			"err", err,
		)

		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		moveShares(ctx, system, deps)
	}()
	wg.Wait()

	if ctx.Err() != nil {
		os.Exit(1)
	}
}
