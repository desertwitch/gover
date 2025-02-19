package main

import (
	"fmt"
	"log/slog"
	"path/filepath"
)

func establishPaths(moveables []*Moveable) ([]*Moveable, error) {
	var filtered []*Moveable

OUTER:
	for _, m := range moveables {
		if err := establishPath(m); err != nil {
			slog.Warn("Skipped job: cannot generate destination path for job", "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}

		for _, h := range m.Hardlinks {
			if err := establishPath(h); err != nil {
				slog.Warn("Skipped job: cannot generate destination path for subjob", "path", h.SourcePath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
				continue OUTER
			}
		}

		for _, s := range m.Symlinks {
			if err := establishPath(s); err != nil {
				slog.Warn("Skipped job: cannot generate destination path for subjob", "path", s.SourcePath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
				continue OUTER
			}
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
		return fmt.Errorf("dir path root is nil")
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
