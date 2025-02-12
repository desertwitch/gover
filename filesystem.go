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

func getMetadata(path string) (*Metadata, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %v", path, err)
	}

	stat := info.Sys().(*syscall.Stat_t)

	metadata := &Metadata{
		Inode:       stat.Ino,
		Permissions: info.Mode().Perm(),
		UID:         stat.Uid,
		GID:         stat.Gid,
		CreatedAt:   stat.Ctim,
		ModifiedAt:  stat.Mtim,
		Size:        stat.Size,
		IsDir:       info.Mode().IsDir(),
	}

	return metadata, nil
}

func walkParentDirs(m *Moveable, basePath string) (map[string]*Metadata, error) {
	path := m.Path
	parentDirs := make(map[string]*Metadata)

	for path != basePath && path != "/" && path != "." {
		path = filepath.Dir(path)

		if strings.HasPrefix(path, basePath) && path != basePath {
			metadata, err := getMetadata(path)
			if err != nil {
				return nil, fmt.Errorf("failed to get metadata for %s: %v", path, err)
			}
			parentDirs[path] = metadata
		} else {
			break
		}
	}

	return parentDirs, nil
}

func isEmptyFolder(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}
