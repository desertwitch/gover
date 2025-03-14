package validation

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/queue"
)

func ValidateMoveables(ctx context.Context, q *queue.EnumerationQueue) error {
	if err := queue.Process(ctx, q, func(m *filesystem.Moveable) bool {
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
	}, true); err != nil {
		return err
	}

	return nil
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
