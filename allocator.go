package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func getMoveables(source UnraidStoreable, share *UnraidShare) ([]*Moveable, error) {
	var moveables []*Moveable

	shareDir := filepath.Join(source.GetFSPath(), share.Name)

	err := filepath.WalkDir(shareDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Println("Error accessing path:", path, err)
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				fmt.Println("Error getting file info:", err)
				return nil
			}

			stat := info.Sys().(*syscall.Stat_t)
			moveable := &Moveable{
				Share:  share,
				Inode:  stat.Ino,
				Size:   stat.Size,
				Path:   path,
				Source: source,
			}
			moveables = append(moveables, moveable)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return moveables, nil
}

// func proposeDestinations(moveables []*Moveable, unraid *UnraidSystem) error {

// }

func allocateDisksBySplitLevel(m *Moveable, maxLevel int) ([]*UnraidDisk, error) {
	foundDisks := []*UnraidDisk{}

	path := filepath.Dir(m.Path)

	relPath, err := filepath.Rel(m.Source.GetFSPath(), path)
	if err != nil {
		return nil, fmt.Errorf("failed deriving subpath: %v", err)
	}

	pathParts := strings.Split(relPath, string(os.PathSeparator))
	splitLevel := len(pathParts)

	if splitLevel == 0 {
		return nil, fmt.Errorf("invalid path with split level of zero: %s", path)
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
				fmt.Printf("Probe: %s\n", dirToCheck)
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
