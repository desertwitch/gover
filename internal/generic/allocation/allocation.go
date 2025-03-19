package allocation

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/desertwitch/gover/internal/generic/configuration"
	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/schema"
)

type fsProvider interface {
	Exists(path string) (bool, error)
	GetDiskUsage(s schema.Storage) (filesystem.DiskStats, error)
	HasEnoughFreeSpace(s schema.Storage, minFree uint64, fileSize uint64) (bool, error)
}

type allocInfo struct {
	sourcePath    string
	sourceBase    string
	allocatedDisk schema.Disk
}

type Handler struct {
	sync.RWMutex
	fsHandler             fsProvider
	alreadyAllocated      map[*schema.Moveable]allocInfo
	alreadyAllocatedSpace map[string]uint64
}

func NewHandler(fsHandler fsProvider) *Handler {
	return &Handler{
		fsHandler:             fsHandler,
		alreadyAllocated:      make(map[*schema.Moveable]allocInfo),
		alreadyAllocatedSpace: make(map[string]uint64),
	}
}

func (a *Handler) AllocateArrayDestination(m *schema.Moveable) bool {
	dest, err := a.allocateArrayDestination(m)
	if err != nil {
		slog.Warn("Skipped job: failed to allocate array destination",
			"err", err,
			"job", m.SourcePath,
			"share", m.Share.GetName(),
		)

		return false
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
		return false
	}

	return true
}

func (a *Handler) allocateArrayDestination(m *schema.Moveable) (schema.Disk, error) {
	includedDisks := m.Share.GetIncludedDisks()

	if m.Share.GetSplitLevel() >= 0 {
		returnDisks, err := a.allocateDisksBySplitLevel(m)
		if err != nil && !errors.Is(err, ErrSplitDoesNotExceedLvl) && !errors.Is(err, ErrNotAllocatable) {
			return nil, fmt.Errorf("(alloc) failed allocating by split level: %w", err)
		}
		if returnDisks != nil {
			includedDisks = returnDisks
		}
	}

	switch allocationMethod := m.Share.GetAllocator(); allocationMethod {
	case configuration.AllocHighWater:
		ret, err := a.allocateHighWater(m, includedDisks)
		if err != nil {
			return nil, fmt.Errorf("(alloc) failed allocating by high water: %w", err)
		}
		a.addAllocated(m, ret)
		a.addAllocatedSpace(ret, m.Metadata.Size)

		return ret, nil

	case configuration.AllocFillUp:
		ret, err := a.allocateFillUp(m, includedDisks)
		if err != nil {
			return nil, fmt.Errorf("(alloc) failed allocating by fillup: %w", err)
		}
		a.addAllocated(m, ret)
		a.addAllocatedSpace(ret, m.Metadata.Size)

		return ret, nil

	case configuration.AllocMostFree:
		ret, err := a.allocateMostFree(m, includedDisks)
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
