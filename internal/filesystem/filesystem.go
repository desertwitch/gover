package filesystem

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/desertwitch/gover/internal/unraid"
)

type Moveable struct {
	Share      *unraid.Share
	Source     unraid.Storeable
	SourcePath string
	Dest       unraid.Storeable
	DestPath   string
	Hardlinks  []*Moveable
	IsHardlink bool
	HardlinkTo *Moveable
	Symlinks   []*Moveable
	IsSymlink  bool
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

type Handler struct {
	OSOps    osProvider
	UnixOps  unixProvider
	FSWalker fsWalker
}

func NewHandler(osOps osProvider, unixOps unixProvider) *Handler {
	return &Handler{
		OSOps:    osOps,
		UnixOps:  unixOps,
		FSWalker: &FileWalker{},
	}
}

func (f *Handler) GetMoveables(share *unraid.Share, src unraid.Storeable, dst unraid.Storeable) ([]*Moveable, error) {
	moveables := []*Moveable{}
	preSelection := []*Moveable{}

	shareDir := filepath.Join(src.GetFSPath(), share.Name)

	err := f.FSWalker.WalkDir(shareDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			slog.Warn("Failure for path during walking of directory tree (skipped)",
				"path", path,
				"err", err,
				"share", share.Name,
			)

			return nil
		}

		isEmptyDir := false
		if d.IsDir() {
			isEmptyDir, err = f.IsEmptyFolder(path)
			if err != nil {
				slog.Warn("Failure checking directory for emptiness during walking of tree (skipped)",
					"path", path,
					"err", err,
					"share", share.Name,
				)

				return nil
			}
		}

		if !d.IsDir() || (d.IsDir() && isEmptyDir && path != shareDir) {
			moveable := &Moveable{
				Share:      share,
				Source:     src,
				SourcePath: path,
				Dest:       dst,
			}

			preSelection = append(preSelection, moveable)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking share: %w", err)
	}

	for _, m := range preSelection {
		metadata, err := f.getMetadata(m.SourcePath)
		if err != nil {
			slog.Warn("Skipped job: failed to get metadata",
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)

			continue
		}
		m.Metadata = metadata

		if err := f.walkParentDirs(m, shareDir); err != nil {
			slog.Warn("Skipped job: failed to get parent folders",
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)

			continue
		}

		moveables = append(moveables, m)
	}

	establishSymlinks(moveables, dst)
	establishHardlinks(moveables, dst)
	moveables = removeInternalLinks(moveables)
	moveables = f.removeInUseFiles(moveables)

	return moveables, nil
}

func (f *Handler) removeInUseFiles(moveables []*Moveable) []*Moveable {
	filtered := []*Moveable{}

	for _, m := range moveables {
		if !m.Metadata.IsDir {
			if inUse, err := f.IsFileInUse(m.SourcePath); err != nil {
				slog.Warn("Skipped job: failed to check if file is in use",
					"err", err,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)

				continue
			} else if inUse {
				slog.Warn("Skipped job: file is in use",
					"err", err,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)

				continue
			}
		}
		filtered = append(filtered, m)
	}

	return filtered
}

func establishSymlinks(moveables []*Moveable, dst unraid.Storeable) {
	realFiles := make(map[string]*Moveable)

	for _, m := range moveables {
		if !m.IsHardlink && !m.Metadata.IsSymlink {
			realFiles[m.SourcePath] = m
		}
	}

	for _, m := range moveables {
		if m.Metadata.IsSymlink {
			if target, exists := realFiles[m.Metadata.SymlinkTo]; exists {
				m.IsSymlink = true
				m.SymlinkTo = target

				m.Dest = dst
				target.Symlinks = append(target.Symlinks, m)
			}
		}
	}
}

func establishHardlinks(moveables []*Moveable, dst unraid.Storeable) {
	inodes := make(map[uint64]*Moveable)
	for _, m := range moveables {
		if target, exists := inodes[m.Metadata.Inode]; exists {
			m.IsHardlink = true
			m.HardlinkTo = target

			m.Dest = dst
			target.Hardlinks = append(target.Hardlinks, m)
		} else {
			inodes[m.Metadata.Inode] = m
		}
	}
}

func removeInternalLinks(moveables []*Moveable) []*Moveable {
	var filtered []*Moveable

	for _, m := range moveables {
		if !m.IsSymlink && !m.IsHardlink {
			filtered = append(filtered, m)
		}
	}

	return filtered
}
