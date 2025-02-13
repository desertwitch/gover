package main

import (
	"log/slog"
)

func main() {
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
		files, err := getMoveables(share.CachePool, share)
		if err != nil {
			slog.Error("failed to get moveables", "share", share.Name, "pool", share.CachePool.Name, "err", err)
			continue
		}
		moveables = append(moveables, files...)
	}

	// Secondary to Primary
	for _, share := range shares {
		if share.UseCache != "prefer" || share.CachePool == nil {
			continue
		}
		if share.CachePool2 == nil {
			// Array to Cache
			for _, disk := range disks {
				files, err := getMoveables(disk, share)
				if err != nil {
					slog.Error("failed to get moveables", "share", share.Name, "disk", disk.Name, "err", err)
					continue
				}
				moveables = append(moveables, files...)
			}
		} else {
			// Cache2 to Cache
			files, err := getMoveables(share.CachePool2, share)
			if err != nil {
				slog.Error("failed to get moveables", "share", share.Name, "pool", share.CachePool2.Name, "err", err)
				continue
			}
			moveables = append(moveables, files...)
		}
	}

	for _, moveable := range moveables {
		if moveable.Share.SplitLevel >= 0 {
			_, err := allocateDisksBySplitLevel(moveable, moveable.Share.SplitLevel)
			if err != nil {
				slog.Error("failed to allocate disk by split level",
					"share", moveable.Share.Name,
					"splitlevel", moveable.Share.SplitLevel,
					"path", moveable.Path,
					"err", err,
				)
			}
		}
	}
}
