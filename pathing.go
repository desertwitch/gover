package main

import (
	"fmt"
	"log/slog"
	"path/filepath"
)

func establishPaths(moveables []*Moveable) ([]*Moveable, error) {
	var filtered []*Moveable

	for _, m := range moveables {
		existsOn, existsPath, err := existsOnStorage(m)
		if err != nil {
			slog.Warn("Skipped job: failed establishing path existence for job", "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}
		if existsOn != nil {
			slog.Warn("Skipped job: destination path already exists for job", "path", existsPath, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}

		if err := establishPath(m); err != nil {
			slog.Warn("Skipped job: cannot set destination path for job", "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}

		hardLinkFailure := false
		for _, h := range m.Hardlinks {
			existsOn, existsPath, err := existsOnStorage(h)
			if err != nil {
				slog.Warn("Skipped job: failed establishing path existence for subjob", "err", err, "job", m.SourcePath, "share", m.Share.Name)
				hardLinkFailure = true
				break
			}
			if existsOn != nil {
				slog.Warn("Skipped job: destination path already exists for subjob", "path", existsPath, "job", m.SourcePath, "share", m.Share.Name)
				hardLinkFailure = true
				break
			}
			if err := establishPath(h); err != nil {
				slog.Warn("Skipped job: cannot set destination path for subjob", "path", h.SourcePath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
				hardLinkFailure = true
				break
			}
		}
		if hardLinkFailure {
			continue
		}

		symlinkFailure := false
		for _, s := range m.Symlinks {
			existsOn, existsPath, err := existsOnStorage(s)
			if err != nil {
				slog.Warn("Skipped job: failed establishing path existence for subjob", "err", err, "job", m.SourcePath, "share", m.Share.Name)
				symlinkFailure = true
				break
			}
			if existsOn != nil {
				slog.Warn("Skipped job: destination path already exists for subjob", "path", existsPath, "job", m.SourcePath, "share", m.Share.Name)
				symlinkFailure = true
				break
			}
			if err := establishPath(s); err != nil {
				slog.Warn("Skipped job: cannot set destination path for subjob", "path", s.SourcePath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
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
		return fmt.Errorf("destination for job is nil")
	}

	relPath, err := filepath.Rel(m.Source.GetFSPath(), m.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to rel path: %w", err)
	}
	m.DestPath = filepath.Join(m.Dest.GetFSPath(), relPath)

	if err := establishRelatedDirPaths(m); err != nil {
		return fmt.Errorf("failed related dir path generation: %w", err)
	}

	return nil
}

func establishRelatedDirPaths(m *Moveable) error {
	if m.RootDir == nil {
		return fmt.Errorf("dir root is nil")
	}

	dir := m.RootDir
	for dir != nil {
		relPath, err := filepath.Rel(m.Source.GetFSPath(), dir.SourcePath)
		if err != nil {
			return fmt.Errorf("failed to rel path: %w", err)
		}
		dir.DestPath = filepath.Join(m.Dest.GetFSPath(), relPath)
		dir = dir.Child
	}

	return nil
}
