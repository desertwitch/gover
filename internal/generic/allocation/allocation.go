package allocation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"sync"

	"github.com/desertwitch/gover/internal/generic/configuration"
	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/schema"
)

type fsProvider interface {
	Exists(path string) (bool, error)
	GetDiskUsage(path string) (filesystem.DiskStats, error)
	HasEnoughFreeSpace(s schema.Storage, minFree uint64, fileSize uint64) (bool, error)
}

type enumerationQueue interface {
	DequeueAndProcessConc(ctx context.Context, maxWorkers int, processFunc func(*schema.Moveable) bool, resetQueueAfter bool) error
}

type allocInfo struct {
	sourcePath    string
	sourceBase    string
	allocatedDisk schema.Disk
}

type Handler struct {
	sync.RWMutex
	fsHandler             fsProvider
	alreadyAllocated      map[*schema.Moveable]*allocInfo
	alreadyAllocatedSpace map[string]uint64
}

func NewHandler(fsHandler fsProvider) *Handler {
	return &Handler{
		fsHandler:             fsHandler,
		alreadyAllocated:      make(map[*schema.Moveable]*allocInfo),
		alreadyAllocatedSpace: make(map[string]uint64),
	}
}

func (a *Handler) AllocateArrayDestinations(ctx context.Context, q enumerationQueue) error {
	if err := q.DequeueAndProcessConc(ctx, runtime.NumCPU(), func(m *schema.Moveable) bool {
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
	}, true); err != nil {
		return err
	}

	return nil
}

func (a *Handler) allocateArrayDestination(m *schema.Moveable) (schema.Disk, error) {
	includedDisks := m.Share.GetIncludedDisks()
	excludedDisks := m.Share.GetExcludedDisks()

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
