package allocation

import (
	"log/slog"
	"sort"

	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/storage"
)

func (a *Handler) allocateFillUp(m *filesystem.Moveable, includedDisks map[string]storage.Disk, excludedDisks map[string]storage.Disk) (storage.Disk, error) {
	diskStats := make(map[string]filesystem.DiskStats)
	disks := []storage.Disk{}

	for name, disk := range includedDisks {
		if _, exists := excludedDisks[name]; exists {
			continue
		}

		stats, err := a.FSOps.GetDiskUsage(disk.GetFSPath())
		if err != nil {
			slog.Warn("Skipped disk for fill-up consideration",
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
		return diskStats[disks[i].GetName()].FreeSpace < diskStats[disks[j].GetName()].FreeSpace
	})

	for _, disk := range disks {
		enoughSpace, err := a.FSOps.HasEnoughFreeSpace(disk, m.Share.GetSpaceFloor(), (a.getAllocatedSpace(disk) + m.Metadata.Size))
		if err != nil {
			slog.Warn("Skipped disk for fill-up consideration",
				"disk", disk.GetName(),
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.GetName(),
			)

			continue
		}
		if enoughSpace && diskStats[disk.GetName()].FreeSpace > m.Share.GetSpaceFloor() {
			return disk, nil
		}
	}

	return nil, ErrNotAllocatable
}
