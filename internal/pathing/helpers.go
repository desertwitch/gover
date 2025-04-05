package pathing

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/desertwitch/gover/internal/schema"
)

// ExistsOnStorage checks if a [schema.Moveable] path exists on the allocated
// storage.
//
// For the allocated destination as a [schema.Pool], it checks if the path
// exists only on that specific [schema.Pool].
//
// For the allocated destination as a [schema.Disk], it checks if the path
// exists on any of the [schema.Share]'s included disks (of an array), to avoid
// duplication when pooled.
func (f *Handler) ExistsOnStorage(m *schema.Moveable) (string, error) {
	if m.Dest == nil {
		return "", ErrNilDestination
	}

	switch dest := m.Dest.(type) {
	case schema.Disk:
		for _, disk := range m.Share.GetIncludedDisks() {
			alreadyExists, existsPath, err := f.existsOnStorageCandidate(m, disk)
			if err != nil {
				return "", err
			}
			if alreadyExists {
				return existsPath, nil
			}
		}

		return "", nil

	case schema.Pool:
		alreadyExists, existsPath, err := f.existsOnStorageCandidate(m, dest)
		if err != nil {
			return "", err
		}
		if alreadyExists {
			return existsPath, nil
		}

		return "", nil

	default:
		return "", ErrImpossibleType
	}
}

// existsOnStorageCandidate checks if a [schema.Moveable] path exists on a
// specific [schema.Storage].
func (f *Handler) existsOnStorageCandidate(m *schema.Moveable, destCandidate schema.Storage) (bool, string, error) {
	relPath, err := filepath.Rel(m.Source.GetFSPath(), m.SourcePath)
	if err != nil {
		return false, "", fmt.Errorf("(fs-existson) failed to rel: %w", err)
	}

	dstPath := filepath.Join(destCandidate.GetFSPath(), relPath)

	if _, err := f.osHandler.Stat(dstPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, "", nil
		}

		return false, "", fmt.Errorf("(fs-existson) failed to stat: %w", err)
	}

	return true, dstPath, nil
}
