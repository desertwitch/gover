package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

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
