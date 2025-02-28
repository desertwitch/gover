package allocation

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
)

func (a *AllocationImpl) AllocateDisksBySplitLevel(m *filesystem.Moveable) (map[string]*unraid.UnraidDisk, error) {
	matches := make(map[int]map[string]*unraid.UnraidDisk)
	splitDoesNotExceedLvl := true

	mainMatches, mainLevel, err := a.findDisksBySplitLevel(m)
	if err != nil {
		if !errors.Is(err, ErrSplitDoesNotExceedLvl) {
			slog.Warn("Skipped job path for split-level consideration", "path", m.SourcePath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
		}
	} else {
		splitDoesNotExceedLvl = false
	}

	if err != nil || len(mainMatches) > 0 {
		matches[mainLevel] = make(map[string]*unraid.UnraidDisk)
		for _, disk := range mainMatches {
			matches[mainLevel][disk.Name] = disk
		}
	}

	if len(m.Hardlinks) > 0 {
		for _, s := range m.Hardlinks {
			subMatches, subLevel, err := a.findDisksBySplitLevel(s)
			if err != nil {
				if !errors.Is(err, ErrSplitDoesNotExceedLvl) {
					slog.Warn("Skipped hardlink for split-level consideration", "path", s.SourcePath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
				}
				continue
			} else {
				splitDoesNotExceedLvl = false
			}

			if len(subMatches) > 0 {
				if matches[subLevel] == nil {
					matches[subLevel] = make(map[string]*unraid.UnraidDisk)
				}
				for _, disk := range subMatches {
					matches[subLevel][disk.Name] = disk
				}
			}
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
	}

	if len(matches[mainLevel]) > 0 {
		return matches[mainLevel], nil
	}

	if splitDoesNotExceedLvl {
		return nil, nil
	}

	return nil, ErrNotAllocatable
}

func (a *AllocationImpl) findDisksBySplitLevel(m *filesystem.Moveable) ([]*unraid.UnraidDisk, int, error) {
	var foundDisks []*unraid.UnraidDisk
	path := filepath.Dir(m.SourcePath)

	relPath, err := filepath.Rel(m.Source.GetFSPath(), path)
	if err != nil {
		return nil, -1, fmt.Errorf("failed deriving subpath: %w", err)
	}

	pathParts := strings.Split(relPath, string(os.PathSeparator))
	splitLevel := len(pathParts)

	if splitLevel == 0 {
		return nil, -1, fmt.Errorf("calc split level of zero: %s", path)
	}

	maxLevel := m.Share.SplitLevel

	if splitLevel <= maxLevel {
		return nil, -1, ErrSplitDoesNotExceedLvl
	} else {
		for i := len(pathParts[maxLevel:]); i > 0; i-- {
			subPath := filepath.Join(pathParts[:maxLevel+i]...)
			found := false
			for name, disk := range m.Share.IncludedDisks {
				if _, exists := m.Share.ExcludedDisks[name]; exists {
					continue
				}
				dirToCheck := filepath.Join(disk.FSPath, subPath)
				if _, err := a.OSOps.Stat(dirToCheck); err == nil {
					enoughSpace, err := a.FSOps.HasEnoughFreeSpace(disk, m.Share.SpaceFloor, m.Metadata.Size)
					if err != nil {
						slog.Warn("Skipped disk for split-level consideration", "disk", name, "err", err, "job", m.SourcePath, "share", m.Share.Name)
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
}
