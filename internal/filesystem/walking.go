package filesystem

import (
	"io/fs"
	"path/filepath"
)

type fsWalker interface {
	WalkDir(root string, fn fs.WalkDirFunc) error
}

type FileWalker struct{}

func (*FileWalker) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}
