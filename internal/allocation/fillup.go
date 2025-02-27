package allocation

import (
	"log/slog"
	"sort"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
)

func allocateFillUpDisk(m *filesystem.Moveable, includedDisks map[string]*unraid.UnraidDisk, excludedDisks map[string]*unraid.UnraidDisk) (*unraid.UnraidDisk, error) {
	diskStats := make(map[*unraid.UnraidDisk]filesystem.DiskStats)
	var disks []*unraid.UnraidDisk

	for name, disk := range includedDisks {
		if _, exists := excludedDisks[name]; exists {
			continue
		}

		stats, err := filesystem.GetDiskUsage(disk.FSPath)
		if err != nil {
			slog.Warn("Skipped disk for fill-up consideration", "disk", disk.Name, "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}
		diskStats[disk] = stats

		disks = append(disks, disk)
	}

	sort.Slice(disks, func(i, j int) bool {
		return diskStats[disks[i]].FreeSpace < diskStats[disks[j]].FreeSpace
	})

	for _, disk := range disks {
		enoughSpace, err := filesystem.HasEnoughFreeSpace(disk, m.Share.SpaceFloor, m.Metadata.Size)
		if err != nil {
			slog.Warn("Skipped disk for fill-up consideration", "disk", disk.Name, "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}
		if enoughSpace && diskStats[disk].FreeSpace > m.Share.SpaceFloor {
			return disk, nil
		}
	}

	return nil, ErrNotAllocatable
}
