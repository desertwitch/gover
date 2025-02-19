package main

import (
	"fmt"
	"log/slog"
)

func validateMoveables(moveables []*Moveable) ([]*Moveable, error) {
	var filtered []*Moveable
	for _, m := range moveables {
		if _, err := validateMoveable(m, make(map[*Moveable]bool)); err != nil {
			slog.Warn("Skipped job: failed pre-move validation", "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}
		filtered = append(filtered, m)
	}
	return filtered, nil
}

func validateMoveable(m *Moveable, visited map[*Moveable]bool) (bool, error) {
	if visited[m] {
		return false, fmt.Errorf("circular reference")
	}
	visited[m] = true
	defer delete(visited, m)

	if m.Share == nil {
		return false, fmt.Errorf("no share information")
	}

	if m.Metadata == nil {
		return false, fmt.Errorf("no metadata")
	}

	if m.RootDir == nil {
		return false, fmt.Errorf("no root dir")
	}

	if m.DeepestDir == nil {
		return false, fmt.Errorf("no deepest dir")
	}

	if m.Source == nil || m.SourcePath == "" {
		return false, fmt.Errorf("no source or source path")
	}

	if m.Dest == nil || m.DestPath == "" {
		return false, fmt.Errorf("no destination or destination path")
	}

	for _, h := range m.Hardlinks {
		if _, err := validateMoveable(h, visited); err != nil {
			return false, err
		}
		if !h.Hardlink {
			return false, fmt.Errorf("hardlink bool is false")
		}
		if h.HardlinkTo == nil {
			return false, fmt.Errorf("no hardlink target")
		}
	}

	for _, s := range m.Symlinks {
		if _, err := validateMoveable(s, visited); err != nil {
			return false, err
		}
		if !s.Symlink {
			return false, fmt.Errorf("symlink bool is false")
		}
		if s.SymlinkTo == nil {
			return false, fmt.Errorf("no symlink target")
		}
	}

	numDirsA := 0
	dirA := m.RootDir
	for dirA != nil {
		if dirA.Metadata == nil {
			return false, fmt.Errorf("no related dir metadata")
		}
		if dirA.Metadata.IsSymlink {
			return false, fmt.Errorf("related dir is a symlink")
		}
		if !dirA.Metadata.IsDir {
			return false, fmt.Errorf("related dir is not a dir")
		}
		if dirA.SourcePath == "" {
			return false, fmt.Errorf("no related dir source path")
		}
		if dirA.DestPath == "" {
			return false, fmt.Errorf("no related dir destination path")
		}
		dirA = dirA.Child
		numDirsA++
	}

	numDirsB := 0
	dirB := m.DeepestDir
	for dirB != nil {
		if dirB.Metadata == nil {
			return false, fmt.Errorf("no related dir metadata")
		}
		if dirB.Metadata.IsSymlink {
			return false, fmt.Errorf("related dir is a symlink")
		}
		if !dirB.Metadata.IsDir {
			return false, fmt.Errorf("related dir is not a dir")
		}
		if dirB.SourcePath == "" {
			return false, fmt.Errorf("no related dir source path")
		}
		if dirB.DestPath == "" {
			return false, fmt.Errorf("no related dir destination path")
		}
		dirB = dirB.Parent
		numDirsB++
	}

	if numDirsA != numDirsB {
		return false, fmt.Errorf("related dir parent/child mismatch")
	}

	return true, nil
}
