package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getMoveables(source UnraidStoreable, share *UnraidShare) ([]*Moveable, error) {
	var moveables []*Moveable

	shareDir := filepath.Join(source.GetFSPath(), share.Name)
	inodes := make(map[uint64]*Moveable)

	err := filepath.WalkDir(shareDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Println("Error accessing path:", path, err)
			return nil
		}

		isEmptyDir := false
		if d.IsDir() {
			isEmptyDir, err = isEmptyFolder(path)
			if err != nil {
				fmt.Println("Error checking for folder emptiness:", path, err)
				return nil
			}
		}

		if !d.IsDir() || (d.IsDir() && isEmptyDir) {
			moveable := &Moveable{
				Share:  share,
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

	for _, m := range moveables {
		metadata, err := getMetadata(m.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to get metadata for %s: %v", m.Path, err)
		}
		m.Metadata = metadata

		// TO-DO SYMLINKS
		if value, exists := inodes[metadata.Inode]; exists {
			m.Hardlink = true
			m.HardlinkTo = value
		} else {
			inodes[metadata.Inode] = m
		}

		parents, err := walkParentDirs(m, shareDir)
		if err != nil {
			return nil, fmt.Errorf("failed to get parents for %s: %v", m.Path, err)
		}
		m.ParentDirs = parents
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
				fmt.Printf("Probe [%s]: %s\n", name, dirToCheck)
				if _, err := os.Stat(dirToCheck); err == nil {
					fmt.Printf("Found disk: %s\n", name)
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
