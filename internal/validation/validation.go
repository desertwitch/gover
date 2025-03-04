package validation

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/desertwitch/gover/internal/filesystem"
)

func ValidateMoveables(moveables []*filesystem.Moveable) ([]*filesystem.Moveable, error) {
	filtered := []*filesystem.Moveable{}

	for _, m := range moveables {
		if err := validateMoveable(m); err != nil {
			slog.Warn("Skipped job: failed pre-move validation for job",
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)

			continue
		}

		hardLinkFailure := false

		for _, h := range m.Hardlinks {
			if err := validateMoveable(h); err != nil {
				slog.Warn("Skipped job: failed pre-move validation for subjob",
					"path", h.SourcePath,
					"err", err,
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
			if err := validateMoveable(s); err != nil {
				slog.Warn("Skipped job: failed pre-move validation for subjob",
					"path", s.SourcePath,
					"err", err,
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

func validateMoveable(m *filesystem.Moveable) error {
	if err := validateBasicAttributes(m); err != nil {
		return err
	}

	if err := validateLinks(m); err != nil {
		return err
	}

	if err := validateDirectories(m); err != nil {
		return err
	}

	return nil
}

func validateBasicAttributes(m *filesystem.Moveable) error {
	if m.Share == nil {
		return ErrNoShareInfo
	}

	if m.Metadata == nil {
		return ErrNoMetadata
	}

	if m.RootDir == nil {
		return ErrNoRootDir
	}

	if m.DeepestDir == nil {
		return ErrNoDeepestDir
	}

	if m.Source == nil || m.SourcePath == "" {
		return ErrNoSource
	}

	if !filepath.IsAbs(m.SourcePath) {
		return ErrSourcePathRelative
	}

	if !strings.HasPrefix(m.SourcePath, m.Source.GetFSPath()) {
		return ErrSourceMismatch
	}

	if m.Dest == nil || m.DestPath == "" {
		return ErrNoDestination
	}

	if !filepath.IsAbs(m.DestPath) {
		return ErrDestPathRelative
	}

	if !strings.HasPrefix(m.DestPath, m.Dest.GetFSPath()) {
		return ErrDestMismatch
	}

	return nil
}

func validateLinks(m *filesystem.Moveable) error {
	if m.IsHardlink {
		if m.HardlinkTo == nil {
			return ErrNoHardlinkTarget
		}

		if m.Hardlinks != nil {
			return ErrHardlinkHasSublinks
		}
	} else if m.HardlinkTo != nil {
		return ErrHardlinkSetTarget
	}

	if m.IsSymlink {
		if m.SymlinkTo == nil {
			return ErrNoSymlinkTarget
		}

		if m.Symlinks != nil {
			return ErrSymlinkHasSublinks
		}
	} else if m.SymlinkTo != nil {
		return ErrSymlinkSetTarget
	}

	return nil
}

func validateDirectories(m *filesystem.Moveable) error {
	numDirsA := 0

	dirA := m.RootDir
	for dirA != nil {
		if err := validateDirectory(dirA); err != nil {
			return err
		}

		dirA = dirA.Child
		numDirsA++
	}

	numDirsB := 0

	dirB := m.DeepestDir
	for dirB != nil {
		if err := validateDirectory(dirB); err != nil {
			return err
		}

		dirB = dirB.Parent
		numDirsB++
	}

	if numDirsA != numDirsB {
		return ErrParentChildMismatch
	}

	if err := validateDirRootConnection(m); err != nil {
		return err
	}

	return nil
}

func validateDirectory(d *filesystem.RelatedDirectory) error {
	if d.Metadata == nil {
		return ErrNoRelatedMetadata
	}

	if d.Metadata.IsSymlink {
		return ErrRelatedDirSymlink
	}

	if !d.Metadata.IsDir {
		return ErrRelatedDirNotDir
	}

	if d.SourcePath == "" {
		return ErrNoRelatedSourcePath
	}

	if !filepath.IsAbs(d.SourcePath) {
		return ErrRelatedSourceRelative
	}

	if d.DestPath == "" {
		return ErrNoRelatedDestPath
	}

	if !filepath.IsAbs(d.DestPath) {
		return ErrDestPathRelative
	}

	return nil
}

func validateDirRootConnection(m *filesystem.Moveable) error {
	shareDirSource := filepath.Join(m.Source.GetFSPath(), m.Share.Name)
	if m.RootDir.SourcePath != shareDirSource {
		return fmt.Errorf("%w: %s != %s", ErrSourceNotConnectBase, shareDirSource, m.RootDir.SourcePath)
	}

	shareDirDest := filepath.Join(m.Dest.GetFSPath(), m.Share.Name)
	if m.RootDir.DestPath != shareDirDest {
		return fmt.Errorf("%w: %s != %s", ErrDestNotConnectBase, shareDirDest, m.RootDir.DestPath)
	}

	return nil
}
