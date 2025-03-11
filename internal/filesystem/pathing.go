package filesystem

import (
	"fmt"
	"log/slog"
	"path/filepath"
)

func (f *Handler) EstablishPaths(moveables []*Moveable) ([]*Moveable, error) {
	filtered := []*Moveable{}

	for _, m := range moveables {
		if err := f.establishElementPath(m); err != nil {
			continue
		}

		hardLinkFailure := false
		for _, h := range m.Hardlinks {
			if err := f.establishSubElementPath(h, m); err != nil {
				hardLinkFailure = true

				break
			}
		}
		if hardLinkFailure {
			continue
		}

		symlinkFailure := false
		for _, s := range m.Symlinks {
			if err := f.establishSubElementPath(s, m); err != nil {
				symlinkFailure = true

				break
			}
		}
		if symlinkFailure {
			continue
		}

		filtered = append(filtered, m)
	}

	return filtered, nil
}

func (f *Handler) establishElementPath(elem *Moveable) error {
	existsPath, err := f.ExistsOnStorage(elem)
	if err != nil {
		slog.Warn("Skipped job: failed establishing path existence",
			"err", err,
			"job", elem.SourcePath,
			"share", elem.Share.Name,
		)

		return err
	}

	// A directory is allowed to exist, that gets handled later in IO.
	if !elem.Metadata.IsDir && existsPath != "" {
		slog.Warn("Skipped job: destination path already exists",
			"path", existsPath,
			"job", elem.SourcePath,
			"share", elem.Share.Name,
		)

		return ErrPathExistsOnDest
	}

	if err := establishPath(elem); err != nil {
		slog.Warn("Skipped job: cannot set destination path",
			"err", err,
			"job", elem.SourcePath,
			"share", elem.Share.Name,
		)

		return err
	}

	return nil
}

func (f *Handler) establishSubElementPath(subelem *Moveable, elem *Moveable) error {
	existsPath, err := f.ExistsOnStorage(subelem)
	if err != nil {
		slog.Warn("Skipped job: failed establishing path existence for subjob",
			"err", err,
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.Name,
		)

		return err
	}
	if existsPath != "" {
		slog.Warn("Skipped job: destination path already exists for subjob",
			"path", existsPath,
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.Name,
		)

		return ErrPathExistsOnDest
	}
	if err := establishPath(subelem); err != nil {
		slog.Warn("Skipped job: cannot set destination path for subjob",
			"path", subelem.SourcePath,
			"err", err,
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.Name,
		)

		return err
	}

	return nil
}

func establishPath(m *Moveable) error {
	if m.Dest == nil {
		return fmt.Errorf("(fs-paths) %w", ErrNilDestination)
	}

	if !filepath.IsAbs(m.SourcePath) {
		return fmt.Errorf("(fs-paths) %w: %s", ErrSourceIsRelative, m.SourcePath)
	}

	relPath, err := filepath.Rel(m.Source.GetFSPath(), m.SourcePath)
	if err != nil {
		return fmt.Errorf("(fs-paths) failed to rel: %w", err)
	}
	m.DestPath = filepath.Join(m.Dest.GetFSPath(), relPath)

	if err := establishRelatedDirPaths(m); err != nil {
		return fmt.Errorf("(fs-paths) failed related dir pathing: %w", err)
	}

	return nil
}

func establishRelatedDirPaths(m *Moveable) error {
	if m.RootDir == nil {
		return fmt.Errorf("(fs-dirpaths) %w", ErrNilDirRoot)
	}

	dir := m.RootDir
	for dir != nil {
		relPath, err := filepath.Rel(m.Source.GetFSPath(), dir.SourcePath)
		if err != nil {
			return fmt.Errorf("(fs-dirpaths) failed to rel: %w", err)
		}
		dir.DestPath = filepath.Join(m.Dest.GetFSPath(), relPath)
		dir = dir.Child
	}

	return nil
}
