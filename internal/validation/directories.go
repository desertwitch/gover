package validation

import (
	"fmt"
	"path/filepath"

	"github.com/desertwitch/gover/internal/filesystem"
)

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

	// Special case: Moveable is an empty share folder (the base).
	// We allow this because no directory relations will be processed (later).
	if m.SourcePath == shareDirSource {
		return nil
	}

	if m.RootDir == nil {
		return fmt.Errorf("(validation) %w: root is nil", ErrSourceNotConnectBase)
	}

	if m.RootDir.SourcePath != shareDirSource {
		return fmt.Errorf("(validation) %w: %s != %s", ErrSourceNotConnectBase, shareDirSource, m.RootDir.SourcePath)
	}

	shareDirDest := filepath.Join(m.Dest.GetFSPath(), m.Share.Name)
	if m.RootDir.DestPath != shareDirDest {
		return fmt.Errorf("(validation) %w: %s != %s", ErrDestNotConnectBase, shareDirDest, m.RootDir.DestPath)
	}

	return nil
}
