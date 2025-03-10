package allocation

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
)

type fsProvider interface {
	Exists(path string) (bool, error)
	GetDiskUsage(path string) (filesystem.DiskStats, error)
	HasEnoughFreeSpace(s unraid.Storeable, minFree uint64, fileSize uint64) (bool, error)
}

type Handler struct {
	sync.RWMutex
	FSOps            fsProvider
	alreadyAllocated map[*unraid.Disk]uint64
}

func NewHandler(fsOps fsProvider) *Handler {
	return &Handler{
		FSOps:            fsOps,
		alreadyAllocated: make(map[*unraid.Disk]uint64),
	}
}

func (a *Handler) AllocateArrayDestinations(moveables []*filesystem.Moveable) ([]*filesystem.Moveable, error) {
	filtered := []*filesystem.Moveable{}

	for _, m := range moveables {
		dest, err := a.AllocateArrayDestination(m)
		if err != nil {
			slog.Warn("Skipped job: failed to allocate array destination",
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)

			continue
		}
		m.Dest = dest

		for _, h := range m.Hardlinks {
			h.Dest = dest
		}

		symlinkFailure := false
		for _, s := range m.Symlinks {
			dest, err := a.AllocateArrayDestination(s)
			if err != nil {
				slog.Warn("Skipped job: failed to allocate array destination for subjob",
					"path", s.SourcePath,
					"err", err,
					"subjob", s.SourcePath,
					"job", m.SourcePath,
					"share", s.Share.Name,
				)
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

func (a *Handler) AllocateArrayDestination(m *filesystem.Moveable) (*unraid.Disk, error) {
	includedDisks := m.Share.IncludedDisks
	excludedDisks := m.Share.ExcludedDisks

	if m.Share.SplitLevel >= 0 {
		returnDisks, err := a.AllocateDisksBySplitLevel(m)
		// TO-DO: Configurable if exceeding, but non allocatable, split-levels should proceed.
		if err != nil && !errors.Is(err, ErrSplitDoesNotExceedLvl) && !errors.Is(err, ErrNotAllocatable) {
			return nil, fmt.Errorf("(alloc) failed allocating by split level: %w", err)
		}
		if returnDisks != nil {
			includedDisks = returnDisks
		}
	}

	switch allocationMethod := m.Share.Allocator; allocationMethod {
	case unraid.AllocHighWater:
		ret, err := a.AllocateHighWaterDisk(m, includedDisks, excludedDisks)
		if err != nil {
			return nil, fmt.Errorf("(alloc) failed allocating by high water: %w", err)
		}
		a.addAlreadyAllocated(ret, m.Metadata.Size)

		return ret, nil

	case unraid.AllocFillUp:
		ret, err := a.AllocateFillUpDisk(m, includedDisks, excludedDisks)
		if err != nil {
			return nil, fmt.Errorf("(alloc) failed allocating by fillup: %w", err)
		}
		a.addAlreadyAllocated(ret, m.Metadata.Size)

		return ret, nil

	case unraid.AllocMostFree:
		ret, err := a.AllocateMostFreeDisk(m, includedDisks, excludedDisks)
		if err != nil {
			return nil, fmt.Errorf("(alloc) failed allocating by mostfree: %w", err)
		}
		a.addAlreadyAllocated(ret, m.Metadata.Size)

		return ret, nil

	default:
		return nil, fmt.Errorf("(alloc) %w", ErrNoAllocationMethod)
	}
}
