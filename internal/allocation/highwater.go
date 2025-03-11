package allocation

import (
	"log/slog"
	"sort"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
)

const (
	highWaterDivisor = 2
)

func (a *Handler) allocateHighWaterDisk(m *filesystem.Moveable, includedDisks map[string]*unraid.Disk, excludedDisks map[string]*unraid.Disk) (*unraid.Disk, error) {
	diskStats := make(map[*unraid.Disk]filesystem.DiskStats)
	disks := []*unraid.Disk{}

	var maxDiskSize uint64

	for name, disk := range includedDisks {
		if _, exists := excludedDisks[name]; exists {
			continue
		}

		stats, err := a.FSOps.GetDiskUsage(disk.FSPath)
		if err != nil {
			slog.Warn("Skipped disk for high-water consideration",
				"disk", disk.Name,
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)

			continue
		}
		diskStats[disk] = stats

		if stats.TotalSize > maxDiskSize {
			maxDiskSize = stats.TotalSize
		}

		disks = append(disks, disk)
	}

	if maxDiskSize == 0 {
		return nil, ErrNoDiskStats
	}

	highWaterMark := maxDiskSize / highWaterDivisor

	for highWaterMark > 0 {
		sort.Slice(disks, func(i, j int) bool {
			return diskStats[disks[i]].FreeSpace < diskStats[disks[j]].FreeSpace
		})
		for _, disk := range disks {
			enoughSpace, err := a.FSOps.HasEnoughFreeSpace(disk, m.Share.SpaceFloor, (a.getAllocatedSpace(disk) + m.Metadata.Size))
			if err != nil {
				slog.Warn("Skipped disk for high-water consideration",
					"disk", disk.Name,
					"err", err,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)

				continue
			}
			if stats, found := diskStats[disk]; found && enoughSpace && stats.FreeSpace >= highWaterMark {
				return disk, nil
			}
		}
		highWaterMark /= highWaterDivisor
	}

	return nil, ErrNotAllocatable
}
