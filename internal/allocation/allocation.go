package allocation

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
)

type fsProvider interface {
	GetDiskUsage(path string) (filesystem.DiskStats, error)
	HasEnoughFreeSpace(s unraid.UnraidStoreable, minFree int64, fileSize int64) (bool, error)
}

type osProvider interface {
	Stat(name string) (os.FileInfo, error)
}

type AllocationImpl struct {
	FSOps fsProvider
	OSOps osProvider
}

func (a AllocationImpl) AllocateArrayDestinations(moveables []*filesystem.Moveable) ([]*filesystem.Moveable, error) {
	var filtered []*filesystem.Moveable

	for _, m := range moveables {
		dest, err := allocateArrayDestination(m, a.FSOps, a.OSOps)
		if err != nil {
			slog.Warn("Skipped job: failed to allocate array destination", "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}
		m.Dest = dest

		for _, h := range m.Hardlinks {
			h.Dest = dest
		}

		symlinkFailure := false
		for _, s := range m.Symlinks {
			dest, err := allocateArrayDestination(s, a.FSOps, a.OSOps)
			if err != nil {
				slog.Warn("Skipped job: failed to allocate array destination for subjob", "path", s.SourcePath, "err", err, "job", m.SourcePath, "share", s.Share.Name)
				symlinkFailure = true
				break
			}
			s.Dest = dest
		}
		if symlinkFailure {
			continue
		}

		filtered = append(filtered, m)
	}

	return filtered, nil
}

func allocateArrayDestination(m *filesystem.Moveable, fsOps fsProvider, osOps osProvider) (*unraid.UnraidDisk, error) {
	includedDisks := m.Share.IncludedDisks
	excludedDisks := m.Share.ExcludedDisks

	if m.Share.SplitLevel >= 0 {
		returnDisks, err := allocateDisksBySplitLevel(m, fsOps, osOps)
		// TO-DO: Configurable, if not found split level files should proceed anyhow
		if err != nil {
			return nil, fmt.Errorf("failed allocating by split level: %w", err)
		}
		if returnDisks != nil {
			includedDisks = returnDisks
		}
	}

	switch allocationMethod := m.Share.Allocator; allocationMethod {
	case unraid.AllocHighWater:
		ret, err := allocateHighWaterDisk(m, includedDisks, excludedDisks, fsOps)
		if err != nil {
			return nil, fmt.Errorf("failed allocating by high water: %w", err)
		}
		return ret, nil

	case unraid.AllocFillUp:
		ret, err := allocateFillUpDisk(m, includedDisks, excludedDisks, fsOps)
		if err != nil {
			return nil, fmt.Errorf("failed allocating by fillup: %w", err)
		}
		return ret, nil

	case unraid.AllocMostFree:
		ret, err := allocateMostFreeDisk(m, includedDisks, excludedDisks, fsOps)
		if err != nil {
			return nil, fmt.Errorf("failed allocating by mostfree: %w", err)
		}
		return ret, nil

	default:
		return nil, fmt.Errorf("no allocation method given in configuration")
	}
}
