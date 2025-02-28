package validation

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/desertwitch/gover/internal/filesystem"
)

func ValidateMoveables(moveables []*filesystem.Moveable) ([]*filesystem.Moveable, error) {
	var filtered []*filesystem.Moveable

	for _, m := range moveables {
		if _, err := validateMoveable(m); err != nil {
			slog.Warn("Skipped job: failed pre-move validation for job", "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}

		hardLinkFailure := false
		for _, h := range m.Hardlinks {
			if _, err := validateMoveable(h); err != nil {
				slog.Warn("Skipped job: failed pre-move validation for subjob", "path", h.SourcePath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
				hardLinkFailure = true
				break
			}
		}
		if hardLinkFailure {
			continue
		}

		symlinkFailure := false
		for _, s := range m.Symlinks {
			if _, err := validateMoveable(s); err != nil {
				slog.Warn("Skipped job: failed pre-move validation for subjob", "path", s.SourcePath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
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

func validateMoveable(m *filesystem.Moveable) (bool, error) {
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

	if !filepath.IsAbs(m.SourcePath) {
		return false, fmt.Errorf("source path is relative")
	}

	if !strings.HasPrefix(m.SourcePath, m.Source.GetFSPath()) {
		return false, fmt.Errorf("source path mismatches source fs element")
	}

	if m.Dest == nil || m.DestPath == "" {
		return false, fmt.Errorf("no destination or destination path")
	}

	if !filepath.IsAbs(m.DestPath) {
		return false, fmt.Errorf("destination path is relative")
	}

	if !strings.HasPrefix(m.DestPath, m.Dest.GetFSPath()) {
		return false, fmt.Errorf("destination path mismatches destination fs element")
	}

	if m.Hardlink {
		if m.HardlinkTo == nil {
			return false, fmt.Errorf("no hardlink target")
		}
		if m.Hardlinks != nil {
			return false, fmt.Errorf("hardlink has sublinks")
		}
	} else {
		if m.HardlinkTo != nil {
			return false, fmt.Errorf("hardlink false, but has set target")
		}
	}

	if m.Symlink {
		if m.SymlinkTo == nil {
			return false, fmt.Errorf("no symlink target")
		}
		if m.Symlinks != nil {
			return false, fmt.Errorf("symlink has sublinks")
		}
	} else {
		if m.SymlinkTo != nil {
			return false, fmt.Errorf("symlink false, but has set target")
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
		if !filepath.IsAbs(dirA.SourcePath) {
			return false, fmt.Errorf("related dir source path is relative")
		}
		if dirA.DestPath == "" {
			return false, fmt.Errorf("no related dir destination path")
		}
		if !filepath.IsAbs(dirA.DestPath) {
			return false, fmt.Errorf("related dir destination path is relative")
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
		if !filepath.IsAbs(dirB.SourcePath) {
			return false, fmt.Errorf("related dir source path is relative")
		}
		if dirB.DestPath == "" {
			return false, fmt.Errorf("no related dir destination path")
		}
		if !filepath.IsAbs(dirB.DestPath) {
			return false, fmt.Errorf("related dir destination path is relative")
		}

		dirB = dirB.Parent
		numDirsB++
	}

	if numDirsA != numDirsB {
		return false, fmt.Errorf("related dir parent/child mismatch")
	}

	shareDirSource := filepath.Join(m.Source.GetFSPath(), m.Share.Name)
	if m.RootDir.SourcePath != shareDirSource {
		return false, fmt.Errorf("related dir root does not connect to share base (source): %s != %s", shareDirSource, m.RootDir.SourcePath)
	}

	shareDirDest := filepath.Join(m.Dest.GetFSPath(), m.Share.Name)
	if m.RootDir.DestPath != shareDirDest {
		return false, fmt.Errorf("related dir root does not connect to share base (dest): %s != %s", shareDirDest, m.RootDir.DestPath)
	}

	return true, nil
}
