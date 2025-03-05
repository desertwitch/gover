package filesystem

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type fsWalker interface {
	WalkDir(root string, fn fs.WalkDirFunc) error
}

type FileWalker struct{}

func (*FileWalker) WalkDir(root string, fn fs.WalkDirFunc) error {
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
			} else {
				m.DeepestDir = thisElement
			}

			prevElement = thisElement
		} else {
			break
		}
	}

	if prevElement == nil {
		return ErrNilDirRoot
	}
	m.RootDir = prevElement

	return nil
}
