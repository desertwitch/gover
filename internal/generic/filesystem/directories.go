package filesystem

import (
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/desertwitch/gover/internal/generic/schema"
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
	var prevElement *schema.RelatedDirectory
	path := m.SourcePath

	for path != basePath && path != "/" && path != "." {
		path = filepath.Dir(path)

		if strings.HasPrefix(path, basePath) {
			thisElement := &schema.RelatedDirectory{
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
