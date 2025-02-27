package filesystem

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/desertwitch/gover/internal/unraid"
	"golang.org/x/sys/unix"
)

type osAdapter interface {
	Stat(name string) (os.FileInfo, error)
	Readlink(name string) (string, error)
	ReadDir(name string) ([]os.DirEntry, error)
}

type unixAdapter interface {
	Statfs(path string, buf *unix.Statfs_t) error
	Lstat(path string, stat *unix.Stat_t) error
}

type RelatedElement interface {
	GetMetadata() *Metadata
	GetSourcePath() string
	GetDestPath() string
}

type FilesystemImpl struct {
	OSCalls   osAdapter
	UnixCalls unixAdapter
}

func (f FilesystemImpl) GetDiskUsage(path string) (DiskStats, error) {
	return getDiskUsage(path, f.UnixCalls)
}

func (f FilesystemImpl) HasEnoughFreeSpace(s unraid.UnraidStoreable, minFree int64, fileSize int64) (bool, error) {
	return hasEnoughFreeSpace(s, minFree, fileSize, f.UnixCalls)
}

func (f FilesystemImpl) IsEmptyFolder(path string) (bool, error) {
	return isEmptyFolder(path, f.OSCalls)
}

func (f FilesystemImpl) ExistsOnStorage(m *Moveable) (storeable unraid.UnraidStoreable, existingAtPath string, err error) {
	return existsOnStorage(m, f.OSCalls)
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

func GetMoveables(source unraid.UnraidStoreable, share *unraid.UnraidShare, knownTarget unraid.UnraidStoreable, osa osAdapter, una unixAdapter) ([]*Moveable, error) {
	var moveables []*Moveable
	var preSelection []*Moveable

	shareDir := filepath.Join(source.GetFSPath(), share.Name)

	err := filepath.WalkDir(shareDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		isEmptyDir := false
		if d.IsDir() {
			isEmptyDir, err = isEmptyFolder(path, osa)
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
		metadata, err := getMetadata(m.SourcePath, osa, una)
		if err != nil {
			slog.Warn("Skipped job: failed to get metadata", "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}
		m.Metadata = metadata

		if err := walkParentDirs(m, shareDir, osa, una); err != nil {
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
