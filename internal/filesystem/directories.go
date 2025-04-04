package filesystem

import (
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/desertwitch/gover/internal/schema"
)

// fileWalker is an implementation that wraps [filepath.WalkDir].
type fileWalker struct{}

// newFileWalker returns a pointer to a new [fileWalker].
func newFileWalker() *fileWalker {
	return &fileWalker{}
}

// WalkDir wraps around the existing [filepath.WalkDir] function.
func (*fileWalker) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}

// establishRelatedDirs uses [Handler.walkParentDirs] to establish
// all parent directories of a [schema.Moveable] up to a given base path.
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

// walkParentDirs walks the parent directories of a [schema.Moveable] until a certain
// base path is reached. Both parent/child relations and metadata of any walked folders
// are established and recorded as part of that tree walking process, for reconstruction.
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
