package allocation

import (
	"log/slog"
	"sort"

	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/schema"
)

const (
	highWaterDivisor = 2
)

func (a *Handler) allocateHighWater(m *schema.Moveable, includedDisks map[string]schema.Disk) (schema.Disk, error) { //nolint:ireturn
	diskStats := make(map[string]filesystem.DiskStats)
	disks := []schema.Disk{}

	var maxDiskSize uint64

	for _, disk := range includedDisks {
		stats, err := a.fsHandler.GetDiskUsage(disk)
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
			enoughSpace, err := a.fsHandler.HasEnoughFreeSpace(disk, m.Share.GetSpaceFloor(), (a.getAllocatedSpace(disk) + m.Metadata.Size))
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
