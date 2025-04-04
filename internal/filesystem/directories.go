package filesystem

import (
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/desertwitch/gover/internal/schema"
)

type fileWalker struct{}

func newFileWalker() *fileWalker {
	return &fileWalker{}
}

func (*fileWalker) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}

func (f *Handler) establishRelatedDirs(m *schema.Moveable, basePath string) error {
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

func (f *Handler) walkParentDirs(m *schema.Moveable, basePath string) error {
	var prevElement *schema.Directory

	path := filepath.Clean(m.SourcePath)
	basePath = filepath.Clean(basePath)

	for {
		parentPath := filepath.Dir(path)

		if parentPath == path {
			break
		}

		if !strings.HasPrefix(parentPath, basePath) {
			break
		}

		path = parentPath

		thisElement := &schema.Directory{
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
	}

	m.RootDir = prevElement

	return nil
}
