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
		return fmt.Errorf("(validation) %w", err)
	}

	if err := validateLinks(m); err != nil {
		return fmt.Errorf("(validation) %w", err)
	}

	if err := validateDirectories(m); err != nil {
		return fmt.Errorf("(validation) %w", err)
	}

	return nil
}

func validateBasicAttributes(m *filesystem.Moveable) error {
	if m.Share == nil {
		return fmt.Errorf("(validation) %w", ErrNoShareInfo)
	}

	if m.Metadata == nil {
		return fmt.Errorf("(validation) %w", ErrNoMetadata)
	}

	if m.RootDir == nil {
		return fmt.Errorf("(validation) %w", ErrNoRootDir)
	}

	if m.DeepestDir == nil {
		return fmt.Errorf("(validation) %w", ErrNoDeepestDir)
	}

	if m.Source == nil || m.SourcePath == "" {
		return fmt.Errorf("(validation) %w", ErrNoSource)
	}

	if !filepath.IsAbs(m.SourcePath) {
		return fmt.Errorf("(validation) %w", ErrSourcePathRelative)
	}

	if !strings.HasPrefix(m.SourcePath, m.Source.GetFSPath()) {
		return fmt.Errorf("(validation) %w", ErrSourceMismatch)
	}

	if m.Dest == nil || m.DestPath == "" {
		return fmt.Errorf("(validation) %w", ErrNoDestination)
	}

	if !filepath.IsAbs(m.DestPath) {
		return fmt.Errorf("(validation) %w", ErrDestPathRelative)
	}

	if !strings.HasPrefix(m.DestPath, m.Dest.GetFSPath()) {
		return fmt.Errorf("(validation) %w", ErrDestMismatch)
	}

	return nil
}

func validateLinks(m *filesystem.Moveable) error {
	if m.IsHardlink {
		if m.HardlinkTo == nil {
			return fmt.Errorf("(validation) %w", ErrNoHardlinkTarget)
		}

		if m.Hardlinks != nil {
			return fmt.Errorf("(validation) %w", ErrHardlinkHasSublinks)
		}
	} else if m.HardlinkTo != nil {
		return fmt.Errorf("(validation) %w", ErrHardlinkSetTarget)
	}

	if m.IsSymlink {
		if m.SymlinkTo == nil {
			return fmt.Errorf("(validation) %w", ErrNoSymlinkTarget)
		}

		if m.Symlinks != nil {
			return fmt.Errorf("(validation) %w", ErrSymlinkHasSublinks)
		}
	} else if m.SymlinkTo != nil {
		return fmt.Errorf("(validation) %w", ErrSymlinkSetTarget)
	}

	return nil
}

func validateDirectories(m *filesystem.Moveable) error {
	dir := m.RootDir

	for dir != nil {
		if err := validateDirectory(dir); err != nil {
			return fmt.Errorf("(validation) %w", err)
		}

		dir = dir.Child
	}

	if err := validateDirRootConnection(m); err != nil {
		return fmt.Errorf("(validation) %w", err)
	}

	return nil
}

func validateDirectory(d *filesystem.RelatedDirectory) error {
	if d.Metadata == nil {
		return fmt.Errorf("(validation) %w", ErrNoRelatedMetadata)
	}

	if d.Metadata.IsSymlink {
		return fmt.Errorf("(validation) %w", ErrRelatedDirSymlink)
	}

	if !d.Metadata.IsDir {
		return fmt.Errorf("(validation) %w", ErrRelatedDirNotDir)
	}

	if d.SourcePath == "" {
		return fmt.Errorf("(validation) %w", ErrNoRelatedSourcePath)
	}

	if !filepath.IsAbs(d.SourcePath) {
		return fmt.Errorf("(validation) %w", ErrRelatedSourceRelative)
	}

	if d.DestPath == "" {
		return fmt.Errorf("(validation) %w", ErrNoRelatedDestPath)
	}

	if !filepath.IsAbs(d.DestPath) {
		return fmt.Errorf("(validation) %w", ErrDestPathRelative)
	}

	return nil
}

func validateDirRootConnection(m *filesystem.Moveable) error {
	shareDirSource := filepath.Join(m.Source.GetFSPath(), m.Share.Name)
	if m.RootDir.SourcePath != shareDirSource {
		return fmt.Errorf("(validation) %w: %s != %s", ErrSourceNotConnectBase, shareDirSource, m.RootDir.SourcePath)
	}

	shareDirDest := filepath.Join(m.Dest.GetFSPath(), m.Share.Name)
	if m.RootDir.DestPath != shareDirDest {
		return fmt.Errorf("(validation) %w: %s != %s", ErrDestNotConnectBase, shareDirDest, m.RootDir.DestPath)
	}

	return nil
}
