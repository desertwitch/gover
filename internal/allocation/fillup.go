package allocation

import (
	"log/slog"
	"sort"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
)

func (a *Allocator) AllocateFillUpDisk(m *filesystem.Moveable, includedDisks map[string]*unraid.Disk, excludedDisks map[string]*unraid.Disk) (*unraid.Disk, error) {
	diskStats := make(map[*unraid.Disk]filesystem.DiskStats)
	disks := []*unraid.Disk{}

	for name, disk := range includedDisks {
		if _, exists := excludedDisks[name]; exists {
			continue
		}

		stats, err := a.FSOps.GetDiskUsage(disk.FSPath)
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
		enoughSpace, err := a.FSOps.HasEnoughFreeSpace(disk, m.Share.SpaceFloor, (a.alreadyAllocated[disk] + m.Metadata.Size))
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
