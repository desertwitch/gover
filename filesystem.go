package main

import (
	"fmt"
	"log/slog"
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
			slog.Error("getMoveables: error accessing path")
			return nil
		}

		isEmptyDir := false
		if d.IsDir() {
			isEmptyDir, err = isEmptyFolder(path)
			if err != nil {
				slog.Error("getMoveables: error checking folder emptiness")
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
		slog.Error("getMoveables: error walking directory")
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	for _, m := range moveables {
		metadata, err := getMetadata(m.Path)
		if err != nil {
			slog.Error("getMoveables: failed to get metadata")
			return nil, fmt.Errorf("failed to get metadata for %s: %w", m.Path, err)
		}
		m.Metadata = metadata

		parents, err := walkParentDirs(m, shareDir)
		if err != nil {
			slog.Error("getMoveables: failed to get parents")
			return nil, fmt.Errorf("failed to get parents for %s: %w", m.Path, err)
		}
		m.ParentDirs = parents
	}

	establishSymlinks(moveables)
	establishHardlinks(moveables)

	moveables = removeInternalLinks(moveables)

	return moveables, nil
}

func establishSymlinks(moveables []*Moveable) {
	realFiles := make(map[string]*Moveable)

	for _, m := range moveables {
		if !m.Hardlink && !m.Metadata.IsSymlink {
			realFiles[m.Path] = m
		}
	}

	for _, m := range moveables {
		if m.Metadata.IsSymlink {
			if target, exists := realFiles[m.Metadata.SymlinkTo]; exists {
				m.Symlink = true
				m.SymlinkTo = target

				target.InternalLinks = append(target.InternalLinks, m)
				slog.Debug("establishSymlinks: found internal symlink")
			}
		}
	}
}

func establishHardlinks(moveables []*Moveable) {
	inodes := make(map[uint64]*Moveable)
	for _, m := range moveables {
		if target, exists := inodes[m.Metadata.Inode]; exists {
			m.Hardlink = true
			m.HardlinkTo = target

			target.InternalLinks = append(target.InternalLinks, m)
			slog.Debug("establishHardlinks: found internal hardlink")
		} else {
			inodes[m.Metadata.Inode] = m
		}
	}
}

func getMetadata(path string) (*Metadata, error) {
	info, err := os.Lstat(path)
	if err != nil {
		slog.Error("getMetadata: failed to lstat")
		return nil, fmt.Errorf("failed to lstat %s: %w", path, err)
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
		IsSymlink:   info.Mode()&os.ModeSymlink != 0,
	}

	if metadata.IsSymlink {
		target, err := os.Readlink(path)
		if err != nil {
			slog.Error("getMetadata: failed to read symlink target")
			return nil, fmt.Errorf("failed to read symlink target for %s: %w", path, err)
		}
		metadata.SymlinkTo = target
		slog.Debug("getMetadata: found fs symlink")
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
				slog.Error("walkParentDirs: failed to get metadata")
				return nil, fmt.Errorf("failed to get metadata for %s: %w", path, err)
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
		slog.Error("isEmptyFolder: failed to read folder")
		return false, err
	}
	return len(entries) == 0, nil
}
