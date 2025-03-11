package filesystem

import (
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
)

type fsWalker interface {
	WalkDir(root string, fn fs.WalkDirFunc) error
}

type fileWalker struct{}

func (*fileWalker) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}

type RelatedDirectory struct {
	SourcePath string
	DestPath   string
	Metadata   *Metadata
	Parent     *RelatedDirectory
	Child      *RelatedDirectory
}

func (d *RelatedDirectory) GetMetadata() *Metadata {
	return d.Metadata
}

func (d *RelatedDirectory) GetSourcePath() string {
	return d.SourcePath
}

func (d *RelatedDirectory) GetDestPath() string {
	return d.DestPath
}

func (f *Handler) establishRelatedDirs(m *Moveable, basePath string) error {
	if err := f.walkParentDirs(m, basePath); err != nil {
		slog.Warn("Skipped job: failed to get parent folders",
			"err", err,
			"job", m.SourcePath,
			"share", m.Share.GetName(),
		)

		return err
	}

	return nil
}

func (f *Handler) walkParentDirs(m *Moveable, basePath string) error {
	var prevElement *RelatedDirectory
	path := m.SourcePath

	for path != basePath && path != "/" && path != "." {
		path = filepath.Dir(path)

		if strings.HasPrefix(path, basePath) {
			thisElement := &RelatedDirectory{
				SourcePath: path,
			}

			metadata, err := f.getMetadata(path)
			if err != nil {
				return fmt.Errorf("(fs-parents) failed to get metadata: %w", err)
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
