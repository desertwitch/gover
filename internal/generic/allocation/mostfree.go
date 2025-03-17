package allocation

import (
	"log/slog"
	"sort"

	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/schema"
)

func (a *Handler) allocateMostFree(m *schema.Moveable, includedDisks map[string]schema.Disk, excludedDisks map[string]schema.Disk) (schema.Disk, error) {
	diskStats := make(map[string]filesystem.DiskStats)
	disks := []schema.Disk{}

	for name, disk := range includedDisks {
		if _, exists := excludedDisks[name]; exists {
			continue
		}

		stats, err := a.fsHandler.GetDiskUsage(disk)
		if err != nil {
			slog.Warn("Skipped disk for most-free consideration",
				"disk", disk.GetName(),
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.GetName(),
			)

			continue
		}
		diskStats[disk.GetName()] = stats

		disks = append(disks, disk)
	}

	sort.Slice(disks, func(i, j int) bool {
		return diskStats[disks[i].GetName()].FreeSpace > diskStats[disks[j].GetName()].FreeSpace
	})

	for _, disk := range disks {
		enoughSpace, err := a.fsHandler.HasEnoughFreeSpace(disk, m.Share.GetSpaceFloor(), (a.getAllocatedSpace(disk) + m.Metadata.Size))
		if err != nil {
			slog.Warn("Skipped disk for most-free consideration",
				"disk", disk.GetName(),
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.GetName(),
			)

			continue
		}
		if enoughSpace {
			return disk, nil
		}
	}

	return nil, ErrNotAllocatable
}
