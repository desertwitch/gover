package allocation

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/desertwitch/gover/internal/configuration"
	"github.com/desertwitch/gover/internal/filesystem"
)

type fsProvider interface {
	Exists(path string) (bool, error)
	GetDiskUsage(path string) (filesystem.DiskStats, error)
	HasEnoughFreeSpace(s filesystem.StorageType, minFree uint64, fileSize uint64) (bool, error)
}

type allocInfo struct {
	sourcePath    string
	sourceBase    string
	allocatedDisk filesystem.DiskType
}

type Handler struct {
	sync.RWMutex
	FSOps                 fsProvider
	alreadyAllocated      map[*filesystem.Moveable]*allocInfo
	alreadyAllocatedSpace map[string]uint64
}

func NewHandler(fsOps fsProvider) *Handler {
	return &Handler{
		FSOps:                 fsOps,
		alreadyAllocated:      make(map[*filesystem.Moveable]*allocInfo),
		alreadyAllocatedSpace: make(map[string]uint64),
	}
}

func (a *Handler) AllocateArrayDestinations(moveables []*filesystem.Moveable) ([]*filesystem.Moveable, error) {
	filtered := []*filesystem.Moveable{}

	for _, m := range moveables {
		dest, err := a.allocateArrayDestination(m)
		if err != nil {
			slog.Warn("Skipped job: failed to allocate array destination",
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.GetName(),
			)

			continue
		}
		m.Dest = dest

		for _, h := range m.Hardlinks {
			h.Dest = dest
		}

		symlinkFailure := false
		for _, s := range m.Symlinks {
			dest, err := a.allocateArrayDestination(s)
			if err != nil {
				slog.Warn("Skipped job: failed to allocate array destination for subjob",
					"path", s.SourcePath,
					"err", err,
					"subjob", s.SourcePath,
					"job", m.SourcePath,
					"share", s.Share.GetName(),
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

func (a *Handler) allocateArrayDestination(m *filesystem.Moveable) (filesystem.DiskType, error) {
	includedDisks := m.Share.GetIncludedDisks()
	excludedDisks := m.Share.GetExcludedDisks()

	if m.Share.GetSplitLevel() >= 0 {
		returnDisks, err := a.allocateDisksBySplitLevel(m)
		// TO-DO: Configurable if exceeding, but non allocatable, split-levels should proceed.
		if err != nil && !errors.Is(err, ErrSplitDoesNotExceedLvl) && !errors.Is(err, ErrNotAllocatable) {
			return nil, fmt.Errorf("(alloc) failed allocating by split level: %w", err)
		}
		if returnDisks != nil {
			includedDisks = returnDisks
		}
	}

	switch allocationMethod := m.Share.GetAllocator(); allocationMethod {
	case configuration.AllocHighWater:
		ret, err := a.allocateHighWater(m, includedDisks, excludedDisks)
		if err != nil {
			return nil, fmt.Errorf("(alloc) failed allocating by high water: %w", err)
		}
		a.addAllocated(m, ret)
		a.addAllocatedSpace(ret, m.Metadata.Size)

		return ret, nil

	case configuration.AllocFillUp:
		ret, err := a.allocateFillUp(m, includedDisks, excludedDisks)
		if err != nil {
			return nil, fmt.Errorf("(alloc) failed allocating by fillup: %w", err)
		}
		a.addAllocated(m, ret)
		a.addAllocatedSpace(ret, m.Metadata.Size)

		return ret, nil

	case configuration.AllocMostFree:
		ret, err := a.allocateMostFree(m, includedDisks, excludedDisks)
		if err != nil {
			return nil, fmt.Errorf("(alloc) failed allocating by mostfree: %w", err)
		}
		a.addAllocated(m, ret)
		a.addAllocatedSpace(ret, m.Metadata.Size)

		return ret, nil

	default:
		return nil, fmt.Errorf("(alloc) %w", ErrNoAllocationMethod)
	}
}
