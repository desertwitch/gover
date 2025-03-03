package main

import (
	"context"
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

func processCurrentSystem(ctx context.Context, handlers *taskHandlers) {
	system, err := handlers.UnraidHandler.EstablishSystem()
	if err != nil {
		slog.Error("failed to establish unraid system", "err", err)

		return
	}

	shares := system.Shares
	disks := system.Array.Disks

	// Primary to Secondary
	for _, share := range shares {
		if ctx.Err() != nil {
			return
		}
		if share.UseCache != "yes" || share.CachePool == nil {
			continue
		}
		if share.CachePool2 == nil {
			// Cache to Array
			files, err := handlers.FSHandler.GetMoveables(share.CachePool, share, nil)
			if err != nil {
				slog.Warn("Skipped share: failed to get jobs", "err", err, "share", share.Name)

				continue
			}
			files, err = handlers.AllocHandler.AllocateArrayDestinations(files)
			if err != nil {
				slog.Warn("Skipped share: failed to allocate jobs", "err", err, "share", share.Name)

				continue
			}
			files, err = handlers.FSHandler.EstablishPaths(files)
			if err != nil {
				slog.Warn("Skipped share: failed to establish paths", "err", err, "share", share.Name)

				continue
			}
			files, err = validation.ValidateMoveables(files)
			if err != nil {
				slog.Warn("Skipped share: failed to validate jobs pre-move", "err", err, "share", share.Name)

				continue
			}
			if err := handlers.IOHandler.ProcessMoveables(ctx, files, &io.ProgressReport{}); err != nil {
				slog.Warn("Skipped share: failed to process jobs", "err", err, "share", share.Name)

				continue
			}
		} else {
			// Cache to Cache2
			files, err := handlers.FSHandler.GetMoveables(share.CachePool, share, share.CachePool2)
			if err != nil {
				slog.Warn("Skipped share: failed to get jobs", "err", err, "share", share.Name)

				continue
			}
			files, err = handlers.FSHandler.EstablishPaths(files)
			if err != nil {
				slog.Warn("Skipped share: failed to establish paths", "err", err, "share", share.Name)

				continue
			}
			files, err = validation.ValidateMoveables(files)
			if err != nil {
				slog.Warn("Skipped share: failed to validate jobs pre-move", "err", err, "share", share.Name)

				continue
			}
			if err := handlers.IOHandler.ProcessMoveables(ctx, files, &io.ProgressReport{}); err != nil {
				slog.Warn("Skipped share: failed to process jobs", "err", err, "share", share.Name)

				continue
			}
		}
	}

	// Secondary to Primary
	for _, share := range shares {
		if ctx.Err() != nil {
			return
		}
		if share.UseCache != "prefer" || share.CachePool == nil {
			continue
		}
		if share.CachePool2 == nil {
			// Array to Cache
			for _, disk := range disks {
				files, err := handlers.FSHandler.GetMoveables(disk, share, share.CachePool)
				if err != nil {
					slog.Warn("Skipped share: failed to get jobs", "err", err, "share", share.Name)

					continue
				}
				files, err = handlers.FSHandler.EstablishPaths(files)
				if err != nil {
					slog.Warn("Skipped share: failed to establish paths", "err", err, "share", share.Name)

					continue
				}
				files, err = validation.ValidateMoveables(files)
				if err != nil {
					slog.Warn("Skipped share: failed to validate jobs pre-move", "err", err, "share", share.Name)

					continue
				}
				if err := handlers.IOHandler.ProcessMoveables(ctx, files, &io.ProgressReport{}); err != nil {
					slog.Warn("Skipped share: failed to process jobs", "err", err, "share", share.Name)

					continue
				}
			}
		} else {
			// Cache2 to Cache
			files, err := handlers.FSHandler.GetMoveables(share.CachePool2, share, share.CachePool)
			if err != nil {
				slog.Warn("Skipped share: failed to get jobs", "err", err, "share", share.Name)

				continue
			}
			files, err = handlers.FSHandler.EstablishPaths(files)
			if err != nil {
				slog.Warn("Skipped share: failed to establish paths", "err", err, "share", share.Name)

				continue
			}
			files, err = validation.ValidateMoveables(files)
			if err != nil {
				slog.Warn("Skipped share: failed to validate jobs pre-move", "err", err, "share", share.Name)

				continue
			}
			if err := handlers.IOHandler.ProcessMoveables(ctx, files, &io.ProgressReport{}); err != nil {
				slog.Warn("Skipped share: failed to process jobs", "err", err, "share", share.Name)

				continue
			}
		}
	}
}

func main() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	w := os.Stderr

	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	osProvider := &filesystem.OS{}
	unixProvider := &filesystem.Unix{}
	cfgProvider := &configuration.GodotenvProvider{}

	configOps := configuration.NewConfigHandler(cfgProvider)
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		processCurrentSystem(ctx, deps)
	}()
	wg.Wait()

	if ctx.Err() != nil {
		os.Exit(1)
	}
}
