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

func (f *Handler) GetMoveables(source unraid.Storeable, share *unraid.Share, knownTarget unraid.Storeable) ([]*Moveable, error) {
	moveables := []*Moveable{}
	preSelection := []*Moveable{}

	shareDir := filepath.Join(source.GetFSPath(), share.Name)

	err := f.FSWalker.WalkDir(shareDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil //nolint: nilerr
		}

		isEmptyDir := false
		if d.IsDir() {
			isEmptyDir, err = f.IsEmptyFolder(path)
			if err != nil {
				return nil //nolint: nilerr
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

	establishSymlinks(moveables, knownTarget)
	establishHardlinks(moveables, knownTarget)
	moveables = removeInternalLinks(moveables)

	return moveables, nil
}

func establishSymlinks(moveables []*Moveable, knownTarget unraid.Storeable) {
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

func establishHardlinks(moveables []*Moveable, knownTarget unraid.Storeable) {
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
