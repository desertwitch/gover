package allocation

import (
	"errors"
	"log/slog"
	"sort"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
)

func (a *Allocator) AllocateHighWaterDisk(m *filesystem.Moveable, includedDisks map[string]*unraid.Disk, excludedDisks map[string]*unraid.Disk) (*unraid.Disk, error) {
	diskStats := make(map[*unraid.Disk]filesystem.DiskStats)
	var disks []*unraid.Disk

	var maxDiskSize int64

	for name, disk := range includedDisks {
		if _, exists := excludedDisks[name]; exists {
			continue
		}

		stats, err := a.FSOps.GetDiskUsage(disk.FSPath)
		if err != nil {
			slog.Warn("Skipped disk for high-water consideration", "disk", disk.Name, "err", err, "job", m.SourcePath, "share", m.Share.Name)

			continue
		}
		diskStats[disk] = stats

		if stats.TotalSize > maxDiskSize {
			maxDiskSize = stats.TotalSize
		}

		disks = append(disks, disk)
	}

	if maxDiskSize == 0 {
		return nil, errors.New("failed getting stats for any disk")
	}

	highWaterMark := maxDiskSize / 2

	for highWaterMark > 0 {
		sort.Slice(disks, func(i, j int) bool {
			return diskStats[disks[i]].FreeSpace < diskStats[disks[j]].FreeSpace
		})
		for _, disk := range disks {
			enoughSpace, err := a.FSOps.HasEnoughFreeSpace(disk, m.Share.SpaceFloor, m.Metadata.Size)
			if err != nil {
				slog.Warn("Skipped disk for high-water consideration", "disk", disk.Name, "err", err, "job", m.SourcePath, "share", m.Share.Name)

				continue
			}
			if stats, found := diskStats[disk]; found && enoughSpace && stats.FreeSpace >= highWaterMark {
				return disk, nil
			}
		}
		highWaterMark /= 2
	}

	return nil, ErrNotAllocatable
}
