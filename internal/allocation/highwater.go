package allocation

import (
	"log/slog"
	"sort"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/storage"
)

const (
	highWaterDivisor = 2
)

func (a *Handler) allocateHighWater(m *filesystem.Moveable, includedDisks map[string]storage.Disk, excludedDisks map[string]storage.Disk) (storage.Disk, error) {
	diskStats := make(map[string]filesystem.DiskStats)
	disks := []storage.Disk{}

	var maxDiskSize uint64

	for name, disk := range includedDisks {
		if _, exists := excludedDisks[name]; exists {
			continue
		}

		stats, err := a.FSOps.GetDiskUsage(disk.GetFSPath())
		if err != nil {
			slog.Warn("Skipped disk for high-water consideration",
				"disk", disk.GetName(),
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.GetName(),
			)

			continue
		}
		diskStats[disk.GetName()] = stats

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
			return diskStats[disks[i].GetName()].FreeSpace < diskStats[disks[j].GetName()].FreeSpace
		})
		for _, disk := range disks {
			enoughSpace, err := a.FSOps.HasEnoughFreeSpace(disk, m.Share.GetSpaceFloor(), (a.getAllocatedSpace(disk) + m.Metadata.Size))
			if err != nil {
				slog.Warn("Skipped disk for high-water consideration",
					"disk", disk.GetName(),
					"err", err,
					"job", m.SourcePath,
					"share", m.Share.GetName(),
				)

				continue
			}
			if stats, found := diskStats[disk.GetName()]; found && enoughSpace && stats.FreeSpace >= highWaterMark {
				return disk, nil
			}
		}
		highWaterMark /= highWaterDivisor
	}

	return nil, ErrNotAllocatable
}
