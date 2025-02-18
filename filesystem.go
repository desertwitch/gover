package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func getMoveables(source UnraidStoreable, share *UnraidShare, knownTarget UnraidStoreable) ([]*Moveable, error) {
	var moveables []*Moveable
	var preSelection []*Moveable

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
				Dest:   knownTarget,
			}

			preSelection = append(preSelection, moveable)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking share (shareDir: %s): %w", shareDir, err)
	}

	for _, m := range preSelection {
		metadata, err := getMetadata(m.Path)
		if err != nil {
			slog.Warn("Skipped job: failed to get metadata", "path", m.Path, "err", err)
			continue
		}
		m.Metadata = metadata

		if err := walkParentDirs(m, shareDir); err != nil {
			slog.Warn("Skipped job: failed to get parent folders", "path", m.Path, "err", err)
			continue
		}

		moveables = append(moveables, m)
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
	var stat unix.Stat_t

	if err := unix.Lstat(path, &stat); err != nil {
		return nil, fmt.Errorf("failed to stat: %w", err)
	}

	metadata := &Metadata{
		Inode:      stat.Ino,
		Perms:      (uint32(stat.Mode) & 0777),
		UID:        stat.Uid,
		GID:        stat.Gid,
		CreatedAt:  stat.Ctim,
		ModifiedAt: stat.Mtim,
		Size:       stat.Size,
		IsDir:      (stat.Mode & unix.S_IFMT) == unix.S_IFDIR,
		IsSymlink:  (stat.Mode & unix.S_IFMT) == unix.S_IFLNK,
	}

	if metadata.IsSymlink {
		target, err := filepath.EvalSymlinks(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read symlink: %w", err)
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
				return fmt.Errorf("failed to relative path (basePath: %s): %w", basePath, err)
			}

			thisElement := &RelatedDirectory{
				Path:         path,
				RelativePath: relativePath,
			}

			metadata, err := getMetadata(path)
			if err != nil {
				return fmt.Errorf("failed to get metadata: %w", err)
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
		return false, fmt.Errorf("failed to readdir: %w", err)
	}
	return len(entries) == 0, nil
}

func getDiskUsage(path string) (DiskStats, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return DiskStats{}, fmt.Errorf("failed to statfs: %w", err)
	}
	stats := DiskStats{
		TotalSize: int64(stat.Blocks) * int64(stat.Bsize),
		FreeSpace: int64(stat.Bavail) * int64(stat.Bsize),
	}
	return stats, nil
}

func hasEnoughFreeSpace(s UnraidStoreable, minFree int64, fileSize int64) (bool, error) {
	if fileSize < 0 {
		return false, fmt.Errorf("invalid file size < 0: %d", fileSize)
	}

	path := s.GetFSPath()

	stats, err := getDiskUsage(path)
	if err != nil {
		return false, fmt.Errorf("failed to get usage: %w", err)
	}

	if stats.TotalSize <= 0 || stats.FreeSpace < 0 {
		return false, fmt.Errorf("invalid stats (TotalSize: %d, FreeSpace: %d)", stats.TotalSize, stats.FreeSpace)
	}

	requiredFree := minFree
	if minFree <= fileSize {
		requiredFree = fileSize
	}

	if stats.FreeSpace > requiredFree {
		return true, nil
	}

	return false, nil
}
