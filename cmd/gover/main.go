package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/desertwitch/gover/internal/allocation"
	"github.com/desertwitch/gover/internal/configuration"
	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/io"
	"github.com/desertwitch/gover/internal/unraid"
	"github.com/desertwitch/gover/internal/validation"
	"github.com/lmittmann/tint"
)

func main() {
	osProvider := &filesystem.OS{}
	unixProvider := &filesystem.Unix{}
	cfgProviderGeneric := &configuration.GodotenvProvider{}

	configOps := configuration.NewConfigHandler(cfgProviderGeneric)
	fsOps := filesystem.NewFileHandler(osProvider, unixProvider)
	unraidOps := unraid.NewUnraidHandler(fsOps, configOps)
	allocOps := allocation.NewAllocator(fsOps)
	ioOps := io.NewIOHandler(allocOps, fsOps, osProvider, unixProvider)

	w := os.Stderr

	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	system, err := unraidOps.EstablishSystem()
	if err != nil {
		slog.Error("failed to establish unraid system", "err", err)
		os.Exit(1)
	}

	shares := system.Shares
	disks := system.Array.Disks

	// Primary to Secondary
	for _, share := range shares {
		if share.UseCache != "yes" || share.CachePool == nil {
			continue
		}
		if share.CachePool2 == nil {
			// Cache to Array
			files, err := fsOps.GetMoveables(share.CachePool, share, nil)
			if err != nil {
				slog.Warn("Skipped share: failed to get jobs", "err", err, "share", share.Name)
				continue
			}
			files, err = allocOps.AllocateArrayDestinations(files)
			if err != nil {
				slog.Warn("Skipped share: failed to allocate jobs", "err", err, "share", share.Name)
				continue
			}
			files, err = fsOps.EstablishPaths(files)
			if err != nil {
				slog.Warn("Skipped share: failed to establish paths", "err", err, "share", share.Name)
				continue
			}
			files, err = validation.ValidateMoveables(files)
			if err != nil {
				slog.Warn("Skipped share: failed to validate jobs pre-move", "err", err, "share", share.Name)
				continue
			}
			if err := ioOps.ProcessMoveables(files, &io.InternalProgressReport{}); err != nil {
				slog.Warn("Skipped share: failed to process jobs", "err", err, "share", share.Name)
				continue
			}
		} else {
			// Cache to Cache2
			files, err := fsOps.GetMoveables(share.CachePool, share, share.CachePool2)
			if err != nil {
				slog.Warn("Skipped share: failed to get jobs", "err", err, "share", share.Name)
				continue
			}
			files, err = fsOps.EstablishPaths(files)
			if err != nil {
				slog.Warn("Skipped share: failed to establish paths", "err", err, "share", share.Name)
				continue
			}
			files, err = validation.ValidateMoveables(files)
			if err != nil {
				slog.Warn("Skipped share: failed to validate jobs pre-move", "err", err, "share", share.Name)
				continue
			}
			if err := ioOps.ProcessMoveables(files, &io.InternalProgressReport{}); err != nil {
				slog.Warn("Skipped share: failed to process jobs", "err", err, "share", share.Name)
				continue
			}
		}
	}

	// Secondary to Primary
	for _, share := range shares {
		if share.UseCache != "prefer" || share.CachePool == nil {
			continue
		}
		if share.CachePool2 == nil {
			// Array to Cache
			for _, disk := range disks {
				files, err := fsOps.GetMoveables(disk, share, share.CachePool)
				if err != nil {
					slog.Warn("Skipped share: failed to get jobs", "err", err, "share", share.Name)
					continue
				}
				files, err = fsOps.EstablishPaths(files)
				if err != nil {
					slog.Warn("Skipped share: failed to establish paths", "err", err, "share", share.Name)
					continue
				}
				files, err = validation.ValidateMoveables(files)
				if err != nil {
					slog.Warn("Skipped share: failed to validate jobs pre-move", "err", err, "share", share.Name)
					continue
				}
				if err := ioOps.ProcessMoveables(files, &io.InternalProgressReport{}); err != nil {
					slog.Warn("Skipped share: failed to process jobs", "err", err, "share", share.Name)
					continue
				}
			}
		} else {
			// Cache2 to Cache
			files, err := fsOps.GetMoveables(share.CachePool2, share, share.CachePool)
			if err != nil {
				slog.Warn("Skipped share: failed to get jobs", "err", err, "share", share.Name)
				continue
			}
			files, err = fsOps.EstablishPaths(files)
			if err != nil {
				slog.Warn("Skipped share: failed to establish paths", "err", err, "share", share.Name)
				continue
			}
			files, err = validation.ValidateMoveables(files)
			if err != nil {
				slog.Warn("Skipped share: failed to validate jobs pre-move", "err", err, "share", share.Name)
				continue
			}
			if err := ioOps.ProcessMoveables(files, &io.InternalProgressReport{}); err != nil {
				slog.Warn("Skipped share: failed to process jobs", "err", err, "share", share.Name)
				continue
			}
		}
	}
}
