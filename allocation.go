package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// func proposeDestinations(moveables []*Moveable, unraid *UnraidSystem) error {

// }

func allocateDisksBySplitLevel(m *Moveable, maxLevel int) ([]*UnraidDisk, error) {
	// TO-DO: Delve into hardlinks and allocate disk by amount of disk matches
	// In case of equal matches - pick random or allocation method?
	// Idea is hard-link original could be not found, but child(s) could be found

	foundDisks := []*UnraidDisk{}
	path := filepath.Dir(m.Path)

	relPath, err := filepath.Rel(m.Source.GetFSPath(), path)
	if err != nil {
		return nil, fmt.Errorf("failed deriving subpath: %w", err)
	}

	pathParts := strings.Split(relPath, string(os.PathSeparator))
	splitLevel := len(pathParts)

	if splitLevel == 0 {
		return nil, fmt.Errorf("calculated split level of zero: %s", path)
	}

	if splitLevel <= maxLevel {
		return nil, nil
	} else {
		for i := len(pathParts[maxLevel:]); i > 0; i-- {
			subPath := filepath.Join(pathParts[:maxLevel+i]...)
			found := false
			for name, disk := range m.Share.IncludedDisks {
				if _, exists := m.Share.ExcludedDisks[name]; exists {
					continue
				}
				dirToCheck := filepath.Join(disk.FSPath, subPath)
				if _, err := os.Stat(dirToCheck); err == nil {
					foundDisks = append(foundDisks, disk)
					found = true
				}
			}
			if found {
				return foundDisks, nil
			}
		}
		return nil, nil
	}
}
