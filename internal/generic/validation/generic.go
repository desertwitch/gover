package validation

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/desertwitch/gover/internal/generic/schema"
)

func validateBasicAttributes(m *schema.Moveable) error {
	if m.Share == nil {
		return fmt.Errorf("(validation) %w", ErrNoShareInfo)
	}

	if m.Metadata == nil {
		return fmt.Errorf("(validation) %w", ErrNoMetadata)
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

func validateLinks(m *schema.Moveable) error {
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
