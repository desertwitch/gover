// Package pathing implements routines for translating abstract filesystem
// elements within [schema.Moveable] into concrete and absolute (destination)
// paths.
package pathing

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/desertwitch/gover/internal/schema"
)

// osProvider defines operating system methods needed for pathing.
type osProvider interface {
	Stat(name string) (os.FileInfo, error)
}

// Handler is the principal implementation for the pathing services.
type Handler struct {
	osHandler osProvider
}

// NewHandler returns a pointer to a new pathing [Handler].
func NewHandler(osHandler osProvider) *Handler {
	return &Handler{
		osHandler: osHandler,
	}
}

// EstablishPath is the principal pathing function that ensures that valid
// destination paths are constructed for a [schema.Moveable]'s set (previously
// allocated) destination [schema.Storage].
func (f *Handler) EstablishPath(m *schema.Moveable) bool {
	if err := f.establishElementPath(m); err != nil {
		return false
	}

	hardLinkFailure := false
	for _, h := range m.Hardlinks {
		if err := f.establishSubElementPath(h, m); err != nil {
			hardLinkFailure = true

			break
		}
	}
	if hardLinkFailure {
		return false
	}

	symlinkFailure := false
	for _, s := range m.Symlinks {
		if err := f.establishSubElementPath(s, m); err != nil {
			symlinkFailure = true

			break
		}
	}
	if symlinkFailure {
		return false
	}

	return true
}

// establishElementPath constructs the destination paths for a "parent"
// [schema.Moveable].
func (f *Handler) establishElementPath(elem *schema.Moveable) error {
	existsPath, err := f.ExistsOnStorage(elem)
	if err != nil {
		slog.Warn("Skipped job: failed establishing path existence",
			"err", err,
			"dst", elem.Dest.GetName(),
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return fmt.Errorf("(pathing) %w", err)
	}

	// A directory is allowed to exist, that gets handled later in IO.
	if !elem.Metadata.IsDir && existsPath != "" {
		slog.Warn("Skipped job: destination path already exists",
			"path", existsPath,
			"dst", elem.Dest.GetName(),
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return fmt.Errorf("(pathing) %w", ErrPathExistsOnDest)
	}

	if err := constructPaths(elem); err != nil {
		slog.Warn("Skipped job: cannot set destination path",
			"err", err,
			"dst", elem.Dest.GetName(),
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return fmt.Errorf("(pathing) %w", err)
	}

	return nil
}

// establishSubElementPath constructs the destination paths for a "child"
// [schema.Moveable] subelement.
func (f *Handler) establishSubElementPath(subelem *schema.Moveable, elem *schema.Moveable) error {
	existsPath, err := f.ExistsOnStorage(subelem)
	if err != nil {
		slog.Warn("Skipped job: failed establishing path existence for subjob",
			"err", err,
			"dst", subelem.Dest.GetName(),
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return fmt.Errorf("(pathing) %w", err)
	}
	if existsPath != "" {
		slog.Warn("Skipped job: destination path already exists for subjob",
			"path", existsPath,
			"dst", subelem.Dest.GetName(),
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return fmt.Errorf("(pathing) %w", ErrPathExistsOnDest)
	}
	if err := constructPaths(subelem); err != nil {
		slog.Warn("Skipped job: cannot set destination path for subjob",
			"err", err,
			"dst", subelem.Dest.GetName(),
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return fmt.Errorf("(pathing) %w", err)
	}

	return nil
}

// constructPaths constructs the destination paths for any [schema.Moveable].
func constructPaths(m *schema.Moveable) error {
	if m.Dest == nil {
		return fmt.Errorf("(pathing) %w", ErrNilDestination)
	}

	if !filepath.IsAbs(m.SourcePath) {
		return fmt.Errorf("(pathing) %w: %s", ErrSourceIsRelative, m.SourcePath)
	}

	relPath, err := filepath.Rel(m.Source.GetFSPath(), m.SourcePath)
	if err != nil {
		return fmt.Errorf("(pathing) failed to rel: %w", err)
	}
	m.DestPath = filepath.Join(m.Dest.GetFSPath(), relPath)

	if err := constructRelatedDirPaths(m); err != nil {
		return fmt.Errorf("(pathing) failed related dir pathing: %w", err)
	}

	return nil
}

// constructRelatedDirPaths constructs the destination paths for the directory
// structure stored inside a [schema.Moveable], for later recreation on the
// target [schema.Storage].
func constructRelatedDirPaths(m *schema.Moveable) error {
	dir := m.RootDir

	for dir != nil {
		relPath, err := filepath.Rel(m.Source.GetFSPath(), dir.SourcePath)
		if err != nil {
			return fmt.Errorf("(pathing-dirs) failed to rel: %w", err)
		}
		dir.DestPath = filepath.Join(m.Dest.GetFSPath(), relPath)
		dir = dir.Child
	}

	return nil
}
