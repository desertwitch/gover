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
	}

	shares := system.Shares
	disks := system.Array.Disks

	moveables := []*Moveable{}

	// Primary to Secondary
	for _, share := range shares {
		if share.UseCache != "yes" || share.CachePool == nil {
			continue
		}
		if share.CachePool2 == nil {
			// Cache to Array
			files, err := getMoveables(share.CachePool, share, nil)
			if err != nil {
				slog.Error("failed to get moveables", "err", err)
				continue
			}
			moveables = append(moveables, files...)
		} else {
			// Cache to Cache2
			files, err := getMoveables(share.CachePool, share, share.CachePool2)
			if err != nil {
				slog.Error("failed to get moveables", "err", err)
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
					slog.Error("failed to get moveables", "err", err)
					continue
				}
				moveables = append(moveables, files...)
			}
		} else {
			// Cache2 to Cache
			files, err := getMoveables(share.CachePool2, share, share.CachePool)
			if err != nil {
				slog.Error("failed to get moveables", "err", err)
				continue
			}
			moveables = append(moveables, files...)
		}
	}

	for _, moveable := range moveables {
		dest, err := proposeArrayDestination(moveable)
		if err != nil {
			fmt.Printf("%s --> error: %v\n", moveable.Path, err)
		} else {
			fmt.Printf("%s --> %s [%v]\n", moveable.Path, dest.Name, moveable.Metadata)
		}
	}
}
