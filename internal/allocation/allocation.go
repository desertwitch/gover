package allocation

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/desertwitch/gover/internal/configuration"
	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/schema"
)

// fsProvider defines external filesystem related methods that are needed for
// allocation.
type fsProvider interface {
	Exists(path string) (bool, error)
	GetDiskUsage(s schema.Storage) (filesystem.DiskStats, error)
	HasEnoughFreeSpace(s schema.Storage, minFree uint64, fileSize uint64) (bool, error)
}

// allocInfo holds information about an allocated [schema.Moveable]. It is meant
// to be passed by value.
type allocInfo struct {
	// The full source path to the respective [schema.Moveable].
	sourcePath string

	// The base path of the source filesystem (e.g. /mnt/disk4).
	sourceBase string

	// The target [schema.Disk] the [schema.Moveable] has been allocated to.
	allocatedDisk schema.Disk
}

// Handler is the principal implementation for the allocation services. It is
// safe for concurrent use on unique, non-concurrent [schema.Moveable].
type Handler struct {
	sync.RWMutex

	// An implementation of [fsProvider] for filesystem-related methods.
	fsHandler fsProvider

	// A map of [schema.Moveable] pointers with information about their
	// allocation.
	alreadyAllocated map[*schema.Moveable]allocInfo

	// The total space that will be taken on a target [schema.Disk].
	alreadyAllocatedSpace map[string]uint64 // map[diskName]uint64
}

// NewHandler returns a pointer to a new allocation [Handler].
func NewHandler(fsHandler fsProvider) *Handler {
	return &Handler{
		fsHandler:             fsHandler,
		alreadyAllocated:      make(map[*schema.Moveable]allocInfo),
		alreadyAllocatedSpace: make(map[string]uint64),
	}
}

// AllocateArrayDestination allocates a [schema.Moveable] and its subelements to
// a [schema.Share]'s included disks (which are typically part of an array).
//
// It is the principal method used allocating given [schema.Moveable], where the
// destination field has not yet been set, to target [schema.Disk] (of an
// array).
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

// allocateArrayDestination provides the allocation logic for allocating a
// single [schema.Moveable] to a [schema.Share]'s included disks (typically part
// of an array). For choice of allocation methods, the [schema.Share]'s
// configuration fields are evaluated.
func (a *Handler) allocateArrayDestination(m *schema.Moveable) (schema.Disk, error) { //nolint:ireturn
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
