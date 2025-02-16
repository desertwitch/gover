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
			return nil
		}

		isEmptyDir := false
		if d.IsDir() {
			isEmptyDir, err = isEmptyFolder(path)
			if err != nil {
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
			return nil, fmt.Errorf("failed to get metadata for %s: %w", m.Path, err)
		}
		m.Metadata = metadata

		if err := walkParentDirs(m, shareDir); err != nil {
			return nil, fmt.Errorf("failed to get parents for %s: %w", m.Path, err)
		}
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

				target.Symlinks = append(target.Symlinks, m)
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

			target.Hardlinks = append(target.Hardlinks, m)
		} else {
			inodes[m.Metadata.Inode] = m
		}
	}
}

func getMetadata(path string) (*Metadata, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to lstat %s: %w", path, err)
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("unexpected type from Sys(): %T", info.Sys())
	}

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
			return nil, fmt.Errorf("failed to read symlink target for %s: %w", path, err)
		}
		metadata.SymlinkTo = target
	}

	return metadata, nil
}

func walkParentDirs(m *Moveable, basePath string) error {
	var prevElement *RelatedDirectory
	path := m.Path

	for path != basePath && path != "/" && path != "." {
		path = filepath.Dir(path)

		if strings.HasPrefix(path, basePath) {
			relativePath, err := filepath.Rel(basePath, path)
			if err != nil {
				return fmt.Errorf("failed to establish relative path with base %s for %s", basePath, path)
			}

			thisElement := &RelatedDirectory{
				Path:         path,
				RelativePath: relativePath,
			}

			metadata, err := getMetadata(path)
			if err != nil {
				return fmt.Errorf("failed to get metadata for %s: %w", path, err)
			}
			thisElement.Metadata = metadata

			if prevElement != nil {
				thisElement.Child = prevElement
				prevElement.Parent = thisElement
			}

			prevElement = thisElement
		} else {
			break
		}
	}
	m.RootDir = prevElement

	return nil
}

func isEmptyFolder(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func getDiskUsage(path string) (DiskStats, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return DiskStats{}, err
	}
	stats := DiskStats{
		TotalSize: int64(stat.Blocks) * int64(stat.Bsize),
		FreeSpace: int64(stat.Bavail) * int64(stat.Bsize),
	}
	return stats, nil
}

func hasEnoughFreeSpace(s UnraidStoreable, minFree int64, fileSize int64) (bool, error) {
	if fileSize < 0 {
		return false, fmt.Errorf("invalid file size: %d", fileSize)
	}

	name := s.GetName()
	path := s.GetFSPath()

	stats, err := getDiskUsage(path)
	if err != nil {
		return false, fmt.Errorf("failed to get disk usage for %s: %w", name, err)
	}

	if stats.TotalSize <= 0 || stats.FreeSpace < 0 {
		return false, fmt.Errorf("invalid disk statistics for %s (TotalSize: %d, FreeSpace: %d)", name, stats.TotalSize, stats.FreeSpace)
	}

	requiredFree := minFree
	if minFree <= 0 {
		requiredFree = fileSize
	}

	if stats.FreeSpace > requiredFree {
		return true, nil
	}

	return false, nil
}
