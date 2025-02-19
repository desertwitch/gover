package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func main() {
	w := os.Stderr

	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	system, err := establishSystem()
	if err != nil {
		slog.Error("failed to establish unraid system", "err", err)
		os.Exit(1)
	}

	shares := system.Shares
	disks := system.Array.Disks

	var moveables []*Moveable

	// Primary to Secondary
	for _, share := range shares {
		if share.UseCache != "yes" || share.CachePool == nil {
			continue
		}
		if share.CachePool2 == nil {
			// Cache to Array
			files, err := getMoveables(share.CachePool, share, nil)
			if err != nil {
				slog.Warn("Skipped share: failed to get jobs", "err", err, "share", share.Name)
				continue
			}
			allocateds, err := allocateArrayDestinations(files)
			if err != nil {
				slog.Warn("Skipped share: failed to allocate jobs", "share", share.Name, "err", err)
				continue
			}
			moveables = append(moveables, allocateds...)
		} else {
			// Cache to Cache2
			files, err := getMoveables(share.CachePool, share, share.CachePool2)
			if err != nil {
				slog.Warn("Skipped share: failed to get jobs", "err", err, "share", share.Name)
				continue
			}
			moveables = append(moveables, files...)
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
				files, err := getMoveables(disk, share, share.CachePool)
				if err != nil {
					slog.Warn("Skipped share: failed to get jobs", "err", err, "share", share.Name)
					continue
				}
				moveables = append(moveables, files...)
			}
		} else {
			// Cache2 to Cache
			files, err := getMoveables(share.CachePool2, share, share.CachePool)
			if err != nil {
				slog.Warn("Skipped share: failed to get jobs", "err", err, "share", share.Name)
				continue
			}
			moveables = append(moveables, files...)
		}
	}

	moveables, _ = establishPaths(moveables)

	for _, m := range moveables {
		fmt.Printf("%s --> %s [%v]\n", m.SourcePath, m.DestPath, m)
		for _, h := range m.Hardlinks {
			fmt.Printf("|- %s --> %s [%v]\n", h.SourcePath, h.DestPath, h)
		}
		for _, s := range m.Symlinks {
			fmt.Printf("|- %s --> %s [%v]\n", s.SourcePath, s.DestPath, s)
		}
		fmt.Println()
	}
}
