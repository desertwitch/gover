package filesystem

import (
	"fmt"
	"log/slog"
	"path/filepath"
)

func (f *Handler) EstablishPaths(moveables []*Moveable) ([]*Moveable, error) {
	filtered := []*Moveable{}

	for _, m := range moveables {
		existsPath, err := f.ExistsOnStorage(m)
		if err != nil {
			slog.Warn("Skipped job: failed establishing path existence",
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)

			continue
		}

		// A directory is allowed to exist, that gets handled later in IO.
		if !m.Metadata.IsDir && existsPath != "" {
			slog.Warn("Skipped job: destination path already exists",
				"path", existsPath,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)

			continue
		}

		if err := establishPath(m); err != nil {
			slog.Warn("Skipped job: cannot set destination path",
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)

			continue
		}

		hardLinkFailure := false
		for _, h := range m.Hardlinks {
			existsPath, err := f.ExistsOnStorage(h)
			if err != nil {
				slog.Warn("Skipped job: failed establishing path existence for subjob",
					"err", err,
					"subjob", h.SourcePath,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)
				hardLinkFailure = true

				break
			}
			if existsPath != "" {
				slog.Warn("Skipped job: destination path already exists for subjob",
					"path", existsPath,
					"subjob", h.SourcePath,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)
				hardLinkFailure = true

				break
			}
			if err := establishPath(h); err != nil {
				slog.Warn("Skipped job: cannot set destination path for subjob",
					"path", h.SourcePath,
					"err", err,
					"subjob", h.SourcePath,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)
				hardLinkFailure = true

				break
			}
		}
		if hardLinkFailure {
			continue
		}

		symlinkFailure := false
		for _, s := range m.Symlinks {
			existsPath, err := f.ExistsOnStorage(s)
			if err != nil {
				slog.Warn("Skipped job: failed establishing path existence for subjob",
					"err", err,
					"subjob", s.SourcePath,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)
				symlinkFailure = true

				break
			}
			if existsPath != "" {
				slog.Warn("Skipped job: destination path already exists for subjob",
					"path", existsPath,
					"subjob", s.SourcePath,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)
				symlinkFailure = true

				break
			}
			if err := establishPath(s); err != nil {
				slog.Warn("Skipped job: cannot set destination path for subjob",
					"path", s.SourcePath,
					"err", err,
					"subjob", s.SourcePath,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)
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
