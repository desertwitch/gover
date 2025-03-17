package pathing

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"

	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/schema"
)

type fsProvider interface {
	ExistsOnStorage(m *schema.Moveable) (string, error)
}

type enumerationQueue interface {
	DequeueAndProcessConc(ctx context.Context, maxWorkers int, processFunc func(*schema.Moveable) int, resetQueueAfter bool) error
}

type Handler struct {
	fsHandler fsProvider
}

func NewHandler(fsHandler fsProvider) *Handler {
	return &Handler{
		fsHandler: fsHandler,
	}
}

func (f *Handler) EstablishPaths(ctx context.Context, q enumerationQueue) error {
	q.DequeueAndProcessConc(ctx, runtime.NumCPU(), func(m *schema.Moveable) int {
		if err := f.establishElementPath(m); err != nil {
			return queue.DecisionSkipped
		}

		hardLinkFailure := false
		for _, h := range m.Hardlinks {
			if err := f.establishSubElementPath(h, m); err != nil {
				hardLinkFailure = true

				break
			}
		}
		if hardLinkFailure {
			return queue.DecisionSkipped
		}

		symlinkFailure := false
		for _, s := range m.Symlinks {
			if err := f.establishSubElementPath(s, m); err != nil {
				symlinkFailure = true

				break
			}
		}
		if symlinkFailure {
			return queue.DecisionSkipped
		}

		return queue.DecisionSuccess
	}, true)

	return nil
}

func (f *Handler) establishElementPath(elem *schema.Moveable) error {
	existsPath, err := f.fsHandler.ExistsOnStorage(elem)
	if err != nil {
		slog.Warn("Skipped job: failed establishing path existence",
			"err", err,
			"dst", elem.Dest.GetName(),
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return err
	}

	// A directory is allowed to exist, that gets handled later in IO.
	if !elem.Metadata.IsDir && existsPath != "" {
		slog.Warn("Skipped job: destination path already exists",
			"path", existsPath,
			"dst", elem.Dest.GetName(),
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return ErrPathExistsOnDest
	}

	if err := establishPath(elem); err != nil {
		slog.Warn("Skipped job: cannot set destination path",
			"err", err,
			"dst", elem.Dest.GetName(),
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return err
	}

	return nil
}

func (f *Handler) establishSubElementPath(subelem *schema.Moveable, elem *schema.Moveable) error {
	existsPath, err := f.fsHandler.ExistsOnStorage(subelem)
	if err != nil {
		slog.Warn("Skipped job: failed establishing path existence for subjob",
			"err", err,
			"dst", subelem.Dest.GetName(),
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return err
	}
	if existsPath != "" {
		slog.Warn("Skipped job: destination path already exists for subjob",
			"path", existsPath,
			"dst", subelem.Dest.GetName(),
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return ErrPathExistsOnDest
	}
	if err := establishPath(subelem); err != nil {
		slog.Warn("Skipped job: cannot set destination path for subjob",
			"err", err,
			"dst", subelem.Dest.GetName(),
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return err
	}

	return nil
}

func establishPath(m *schema.Moveable) error {
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

func establishRelatedDirPaths(m *schema.Moveable) error {
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
