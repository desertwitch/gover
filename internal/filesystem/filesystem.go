package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/desertwitch/gover/internal/unraid"
	"golang.org/x/sys/unix"
)

type osProvider interface {
	Stat(name string) (os.FileInfo, error)
	Readlink(name string) (string, error)
	ReadDir(name string) ([]os.DirEntry, error)
}

type unixProvider interface {
	Statfs(path string, buf *unix.Statfs_t) error
	Lstat(path string, stat *unix.Stat_t) error
}

type Moveable struct {
	Share      *unraid.UnraidShare
	Source     unraid.UnraidStoreable
	SourcePath string
	Dest       unraid.UnraidStoreable
	DestPath   string
	Hardlinks  []*Moveable
	Hardlink   bool
	HardlinkTo *Moveable
	Symlinks   []*Moveable
	Symlink    bool
	SymlinkTo  *Moveable
	Metadata   *Metadata
	RootDir    *RelatedDirectory
	DeepestDir *RelatedDirectory
}

func (m *Moveable) GetMetadata() *Metadata {
	return m.Metadata
}

func (m *Moveable) GetSourcePath() string {
	return m.SourcePath
}

func (m *Moveable) GetDestPath() string {
	return m.DestPath
}

type DiskStats struct {
	TotalSize int64
	FreeSpace int64
}

type FilesystemImpl struct {
	OSOps   osProvider
	UnixOps unixProvider
}

func (f FilesystemImpl) GetMoveables(source unraid.UnraidStoreable, share *unraid.UnraidShare, knownTarget unraid.UnraidStoreable) ([]*Moveable, error) {
	var moveables []*Moveable
	var preSelection []*Moveable

	shareDir := filepath.Join(source.GetFSPath(), share.Name)

	err := filepath.WalkDir(shareDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		isEmptyDir := false
		if d.IsDir() {
			isEmptyDir, err = f.IsEmptyFolder(path)
			if err != nil {
				return nil
			}
		}

		if !d.IsDir() || (d.IsDir() && isEmptyDir && path != shareDir) {
			moveable := &Moveable{
				Share:      share,
				Source:     source,
				SourcePath: path,
				Dest:       knownTarget,
			}

			preSelection = append(preSelection, moveable)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking share: %w", err)
	}

	for _, m := range preSelection {
		metadata, err := getMetadata(m.SourcePath, f.OSOps, f.UnixOps)
		if err != nil {
			slog.Warn("Skipped job: failed to get metadata", "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}
		m.Metadata = metadata

		if err := walkParentDirs(m, shareDir, f.OSOps, f.UnixOps); err != nil {
			slog.Warn("Skipped job: failed to get parent folders", "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}

		moveables = append(moveables, m)
	}

	establishSymlinks(moveables, knownTarget)
	establishHardlinks(moveables, knownTarget)
	moveables = removeInternalLinks(moveables)

	return moveables, nil
}

func (f FilesystemImpl) GetDiskUsage(path string) (DiskStats, error) {
	var stat unix.Statfs_t
	if err := f.UnixOps.Statfs(path, &stat); err != nil {
		return DiskStats{}, fmt.Errorf("failed to statfs: %w", err)
	}
	stats := DiskStats{
		TotalSize: int64(stat.Blocks) * int64(stat.Bsize),
		FreeSpace: int64(stat.Bavail) * int64(stat.Bsize),
	}
	return stats, nil
}

func (f FilesystemImpl) HasEnoughFreeSpace(s unraid.UnraidStoreable, minFree int64, fileSize int64) (bool, error) {
	if fileSize < 0 {
		return false, fmt.Errorf("invalid file size < 0: %d", fileSize)
	}

	path := s.GetFSPath()

	stats, err := f.GetDiskUsage(path)
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

func (f FilesystemImpl) IsEmptyFolder(path string) (bool, error) {
	entries, err := f.OSOps.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("failed to readdir: %w", err)
	}
	return len(entries) == 0, nil
}

func (f FilesystemImpl) ExistsOnStorage(m *Moveable) (storeable unraid.UnraidStoreable, existingAtPath string, err error) {
	if m.Dest == nil {
		return nil, "", fmt.Errorf("destination is nil")
	}

	if _, ok := m.Dest.(*unraid.UnraidDisk); ok {
		for name, disk := range m.Share.IncludedDisks {
			if _, exists := m.Share.ExcludedDisks[name]; exists {
				continue
			}
			alreadyExists, existsPath, err := existsOnStorageCandidate(m, disk, f.OSOps)
			if err != nil {
				return nil, "", err
			}
			if alreadyExists {
				return disk, existsPath, nil
			}
		}
		return nil, "", nil
	}

	if pool, ok := m.Dest.(*unraid.UnraidPool); ok {
		alreadyExists, existsPath, err := existsOnStorageCandidate(m, pool, f.OSOps)
		if err != nil {
			return nil, "", err
		}
		if alreadyExists {
			return pool, existsPath, nil
		}
		return nil, "", nil
	}

	return nil, "", fmt.Errorf("impossible storeable type")
}

func existsOnStorageCandidate(m *Moveable, destCandidate unraid.UnraidStoreable, osOps osProvider) (exists bool, existingAtPath string, err error) {
	relPath, err := filepath.Rel(m.Source.GetFSPath(), m.SourcePath)
	if err != nil {
		return false, "", fmt.Errorf("failed to rel path: %w", err)
	}

	dstPath := filepath.Join(destCandidate.GetFSPath(), relPath)

	if _, err := osOps.Stat(dstPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check existence: %w", err)
	}

	return true, dstPath, nil
}

func establishSymlinks(moveables []*Moveable, knownTarget unraid.UnraidStoreable) {
	realFiles := make(map[string]*Moveable)

	for _, m := range moveables {
		if !m.Hardlink && !m.Metadata.IsSymlink {
			realFiles[m.SourcePath] = m
		}
	}

	for _, m := range moveables {
		if m.Metadata.IsSymlink {
			if target, exists := realFiles[m.Metadata.SymlinkTo]; exists {
				m.Symlink = true
				m.SymlinkTo = target

				m.Dest = knownTarget
				target.Symlinks = append(target.Symlinks, m)
			}
		}
	}
}

func establishHardlinks(moveables []*Moveable, knownTarget unraid.UnraidStoreable) {
	inodes := make(map[uint64]*Moveable)
	for _, m := range moveables {
		if target, exists := inodes[m.Metadata.Inode]; exists {
			m.Hardlink = true
			m.HardlinkTo = target

			m.Dest = knownTarget
			target.Hardlinks = append(target.Hardlinks, m)
		} else {
			inodes[m.Metadata.Inode] = m
		}
	}
}

func removeInternalLinks(moveables []*Moveable) []*Moveable {
	var ms []*Moveable

	for _, m := range moveables {
		if !m.Symlink && !m.Hardlink {
			ms = append(ms, m)
		}
	}

	return ms
}
