// Package validation implements routines for validation of [schema.Moveable].
package validation

import (
	"fmt"
	"log/slog"

	"github.com/desertwitch/gover/internal/schema"
)

// ValidateMoveable is the principal function to validate a [schema.Moveable]
// and its subelements.
func ValidateMoveable(m *schema.Moveable) bool {
	if err := validateMoveable(m); err != nil {
		slog.Warn("Skipped job: failed pre-move validation",
			"err", err,
			"job", m.SourcePath,
			"share", m.Share.GetName(),
		)

		return false
	}

	hardLinkFailure := false

	for _, h := range m.Hardlinks {
		if err := validateMoveable(h); err != nil {
			slog.Warn("Skipped job: failed pre-move validation for subjob",
				"path", h.SourcePath,
				"err", err,
				"subjob", h.SourcePath,
				"job", m.SourcePath,
				"share", m.Share.GetName(),
			)

			hardLinkFailure = true

			break
		}
	}

	if hardLinkFailure {
		return false
	}

	symlinkFailure := false

	for _, s := range m.Symlinks {
		if err := validateMoveable(s); err != nil {
			slog.Warn("Skipped job: failed pre-move validation for subjob",
				"path", s.SourcePath,
				"err", err,
				"subjob", s.SourcePath,
				"job", m.SourcePath,
				"share", m.Share.GetName(),
			)

			symlinkFailure = true

			break
		}
	}

	if symlinkFailure {
		return false
	}

	return true
}

// validateMoveable is the principal function validate a single
// [schema.Moveable].
func validateMoveable(m *schema.Moveable) error {
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
