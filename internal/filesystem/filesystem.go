package filesystem

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type Moveable struct {
	Share      ShareType
	Source     StorageType
	SourcePath string
	Dest       StorageType
	DestPath   string
	Hardlinks  []*Moveable
	IsHardlink bool
	HardlinkTo *Moveable
	Symlinks   []*Moveable
	IsSymlink  bool
	SymlinkTo  *Moveable
	Metadata   *Metadata
	RootDir    *RelatedDirectory
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
		FSWalker: &fileWalker{},
	}
}

func (f *Handler) GetMoveables(share ShareType, src StorageType, dst StorageType) ([]*Moveable, error) {
	moveables := []*Moveable{}
	filtered := []*Moveable{}

	shareDir := filepath.Join(src.GetFSPath(), share.GetName())

	err := f.FSWalker.WalkDir(shareDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if path != shareDir {
				slog.Warn("Failure for path during walking of directory tree (was skipped)",
					"path", path,
					"err", err,
					"share", share.GetName(),
				)
			}

			return nil
		}

		isEmptyDir := false
		if d.IsDir() {
			isEmptyDir, err = f.IsEmptyFolder(path)
			if err != nil {
				slog.Warn("Failure checking for emptiness during walking of directory tree (was skipped)",
					"path", path,
					"err", err,
					"share", share.GetName(),
				)

				return nil
			}
		}

		if !d.IsDir() || (d.IsDir() && isEmptyDir) {
			moveable := &Moveable{
				Share:      share,
				Source:     src,
				SourcePath: path,
				Dest:       dst,
			}

			moveables = append(moveables, moveable)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("(fs) error walking %s: %w", shareDir, err)
	}

	for _, m := range moveables {
		if err := f.establishMetadata(m); err != nil {
			continue
		}
		if err := f.establishRelatedDirs(m, shareDir); err != nil {
			continue
		}
		filtered = append(filtered, m)
	}

	establishSymlinks(filtered, dst)
	establishHardlinks(filtered, dst)
	filtered = removeInternalLinks(filtered)
	filtered = f.removeInUseFiles(filtered)

	return filtered, nil
}

func (f *Handler) removeInUseFiles(moveables []*Moveable) []*Moveable {
	filtered := []*Moveable{}

	for _, m := range moveables {
		if !m.Metadata.IsDir {
			if inUse, err := f.IsFileInUse(m.SourcePath); err != nil {
				slog.Warn("Skipped job: failed to check if file is in use",
					"err", err,
					"job", m.SourcePath,
					"share", m.Share.GetName(),
				)

				continue
			} else if inUse {
				slog.Warn("Skipped job: source file is in use",
					"err", err,
					"job", m.SourcePath,
					"share", m.Share.GetName(),
				)

				continue
			}
		}
		filtered = append(filtered, m)
	}

	return filtered
}
