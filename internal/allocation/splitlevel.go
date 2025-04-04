package allocation

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/desertwitch/gover/internal/schema"
)

// allocateDisksBySplitLevel provides the allocation logic for the split-level allocation method.
// It chooses a suitable [schema.Disk] if the source path exceeds a pre-defined split-level threshold
// and is already existing on a [schema.Disk] at a certain split-level. A best-match mechanism considers
// the main [schema.Moveable] and also its subelements, picking the deepest matching available split-path.
func (a *Handler) allocateDisksBySplitLevel(m *schema.Moveable) (map[string]schema.Disk, error) {
	matches := make(map[int]map[string]schema.Disk)
	splitExceedLvl := false

	mainMatches, mainLevel, err := a.findDisksBySplitLevel(m)
	if err != nil {
		if !errors.Is(err, ErrSplitDoesNotExceedLvl) {
			slog.Warn("Skipped job path for split-level consideration",
				"path", m.SourcePath,
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.GetName(),
			)
		}
	} else {
		splitExceedLvl = true

		if len(mainMatches) > 0 {
			matches[mainLevel] = make(map[string]schema.Disk)
			for _, disk := range mainMatches {
				matches[mainLevel][disk.GetName()] = disk
			}
		}
	}

	if len(m.Hardlinks) > 0 { //nolint:nestif
		for _, s := range m.Hardlinks {
			subMatches, subLevel, err := a.findDisksBySplitLevel(s)
			if err != nil {
				if !errors.Is(err, ErrSplitDoesNotExceedLvl) {
					slog.Warn("Skipped hardlink for split-level consideration",
						"path", s.SourcePath,
						"err", err,
						"job", m.SourcePath,
						"share", m.Share.GetName(),
					)
				}
			} else {
				splitExceedLvl = true

				if len(subMatches) > 0 {
					if matches[subLevel] == nil {
						matches[subLevel] = make(map[string]schema.Disk)
					}
					for _, disk := range subMatches {
						matches[subLevel][disk.GetName()] = disk
					}
				}
			}
		}
	}

	if !splitExceedLvl {
		return nil, ErrSplitDoesNotExceedLvl
	}

	maxKey := -1
	for key := range matches {
		if key > maxKey {
			maxKey = key
		}
	}

	if maxKey >= 0 {
		if bestMatch, exists := matches[maxKey]; exists {
			return bestMatch, nil
		}
	}

	return nil, ErrNotAllocatable
}

// findDisksBySplitLevel provides the allocation logic for allocating a single
// [schema.Moveable] according to a configured split-level threshold, if possible.
// It returns a slice of [schema.Disk] that will contain all disks found at the deepest
// possible split-level and also an integer of what that split-level was, for later sorting.
// Both all physically existing paths and paths pre-allocated by the [Handler] are considered.
func (a *Handler) findDisksBySplitLevel(m *schema.Moveable) ([]schema.Disk, int, error) {
	var foundDisks []schema.Disk

	path := filepath.Dir(m.SourcePath)

	relPath, err := filepath.Rel(m.Source.GetFSPath(), path)
	if err != nil {
		return nil, -1, fmt.Errorf("(alloc-splitlvl) failed to rel %w", err)
	}

	pathParts := strings.Split(relPath, string(os.PathSeparator))
	splitLevel := len(pathParts)

	if splitLevel == 0 {
		return nil, -1, fmt.Errorf("(alloc-splitlvl) %w: %s", ErrCalcSplitLvlZero, path)
	}

	maxLevel := m.Share.GetSplitLevel()

	if splitLevel <= maxLevel {
		return nil, -1, fmt.Errorf("(alloc-splitlvl) %w: %d < %d", ErrSplitDoesNotExceedLvl, splitLevel, maxLevel)
	}

	for i := len(pathParts[maxLevel:]); i > 0; i-- {
		subPath := filepath.Join(pathParts[:maxLevel+i]...)
		found := false

		alreadyAllocated := a.getAllocatedsForSubpath(subPath)

		for name, disk := range m.Share.GetIncludedDisks() {
			dirToCheck := filepath.Join(disk.GetFSPath(), subPath)

			existsPhysical, err := a.fsHandler.Exists(dirToCheck)
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				slog.Warn("Skipped disk for split-level consideration",
					"disk", name,
					"err", err,
					"job", m.SourcePath,
					"share", m.Share.GetName(),
				)

				continue
			}

			_, existsAllocated := alreadyAllocated[disk.GetName()]

			if existsPhysical || existsAllocated {
				enoughSpace, err := a.fsHandler.HasEnoughFreeSpace(disk, m.Share.GetSpaceFloor(), (a.getAllocatedSpace(disk) + m.Metadata.Size))
				if err != nil {
					slog.Warn("Skipped disk for split-level consideration",
						"disk", name,
						"err", err,
						"job", m.SourcePath,
						"share", m.Share.GetName(),
					)

					continue
				}
				if enoughSpace {
					foundDisks = append(foundDisks, disk)
					found = true
				}
			}
		}

		if found {
			return foundDisks, i, nil
		}
	}

	return nil, -1, nil
}
