package pathing

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/queue"
)

type fsProvider interface {
	ExistsOnStorage(m *filesystem.Moveable) (string, error)
}

type Handler struct {
	fsHandler fsProvider
}

func NewHandler(fsHandler fsProvider) *Handler {
	return &Handler{
		fsHandler: fsHandler,
	}
}

func (f *Handler) EstablishPaths(ctx context.Context, q *queue.EnumerationQueue) error {
	queue.Process(ctx, q, func(m *filesystem.Moveable) bool {
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
	}, true)

	return nil
}

func (f *Handler) establishElementPath(elem *filesystem.Moveable) error {
	existsPath, err := f.fsHandler.ExistsOnStorage(elem)
	if err != nil {
		slog.Warn("Skipped job: failed establishing path existence",
			"err", err,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return err
	}

	// A directory is allowed to exist, that gets handled later in IO.
	if !elem.Metadata.IsDir && existsPath != "" {
		slog.Warn("Skipped job: destination path already exists",
			"path", existsPath,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return ErrPathExistsOnDest
	}

	if err := establishPath(elem); err != nil {
		slog.Warn("Skipped job: cannot set destination path",
			"err", err,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return err
	}

	return nil
}

func (f *Handler) establishSubElementPath(subelem *filesystem.Moveable, elem *filesystem.Moveable) error {
	existsPath, err := f.fsHandler.ExistsOnStorage(subelem)
	if err != nil {
		slog.Warn("Skipped job: failed establishing path existence for subjob",
			"err", err,
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return err
	}
	if existsPath != "" {
		slog.Warn("Skipped job: destination path already exists for subjob",
			"path", existsPath,
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return ErrPathExistsOnDest
	}
	if err := establishPath(subelem); err != nil {
		slog.Warn("Skipped job: cannot set destination path for subjob",
			"path", subelem.SourcePath,
			"err", err,
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return err
	}

	return nil
}

func establishPath(m *filesystem.Moveable) error {
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

	if err := establishRelatedDirPaths(m); err != nil {
		return fmt.Errorf("(pathing) failed related dir pathing: %w", err)
	}

	return nil
}

func establishRelatedDirPaths(m *filesystem.Moveable) error {
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
