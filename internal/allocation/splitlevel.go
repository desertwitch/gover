package allocation

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
)

func (a *Handler) AllocateDisksBySplitLevel(m *filesystem.Moveable) (map[string]*unraid.Disk, error) {
	matches := make(map[int]map[string]*unraid.Disk)
	splitExceedLvl := false

	mainMatches, mainLevel, err := a.findDisksBySplitLevel(m)
	if err != nil {
		if !errors.Is(err, ErrSplitDoesNotExceedLvl) {
			slog.Warn("Skipped job path for split-level consideration",
				"path", m.SourcePath,
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)
		}
	} else {
		splitExceedLvl = true

		if len(mainMatches) > 0 {
			matches[mainLevel] = make(map[string]*unraid.Disk)
			for _, disk := range mainMatches {
				matches[mainLevel][disk.Name] = disk
			}
		}
	}

	if len(m.Hardlinks) > 0 {
		for _, s := range m.Hardlinks {
			subMatches, subLevel, err := a.findDisksBySplitLevel(s)
			if err != nil {
				if !errors.Is(err, ErrSplitDoesNotExceedLvl) {
					slog.Warn("Skipped hardlink for split-level consideration",
						"path", s.SourcePath,
						"err", err,
						"job", m.SourcePath,
						"share", m.Share.Name,
					)
				}
			} else {
				splitExceedLvl = true

				if len(subMatches) > 0 {
					if matches[subLevel] == nil {
						matches[subLevel] = make(map[string]*unraid.Disk)
					}
					for _, disk := range subMatches {
						matches[subLevel][disk.Name] = disk
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

func (a *Handler) findDisksBySplitLevel(m *filesystem.Moveable) ([]*unraid.Disk, int, error) {
	var foundDisks []*unraid.Disk
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

	maxLevel := m.Share.SplitLevel

	if splitLevel <= maxLevel {
		return nil, -1, fmt.Errorf("(alloc-splitlvl) %w: %d < %d", ErrSplitDoesNotExceedLvl, splitLevel, maxLevel)
	}

	for i := len(pathParts[maxLevel:]); i > 0; i-- {
		subPath := filepath.Join(pathParts[:maxLevel+i]...)
		found := false
		for name, disk := range m.Share.IncludedDisks {
			if _, exists := m.Share.ExcludedDisks[name]; exists {
				continue
			}
			dirToCheck := filepath.Join(disk.FSPath, subPath)
			if exists, err := a.FSOps.Exists(dirToCheck); err == nil && exists {
				enoughSpace, err := a.FSOps.HasEnoughFreeSpace(disk, m.Share.SpaceFloor, (a.alreadyAllocated[disk] + m.Metadata.Size))
				if err != nil {
					slog.Warn("Skipped disk for split-level consideration",
						"disk", name,
						"err", err,
						"job", m.SourcePath,
						"share", m.Share.Name,
					)

					continue
				}
				if enoughSpace {
					foundDisks = append(foundDisks, disk)
					found = true
				}
			} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
				slog.Warn("Skipped disk for split-level consideration",
					"disk", name,
					"err", err,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)

				continue
			}
		}
		if found {
			return foundDisks, i, nil
		}
	}

	return nil, -1, nil
}
