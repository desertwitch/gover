package validation

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/schema"
)

type enumerationQueue interface {
	DequeueAndProcess(ctx context.Context, processFunc func(*schema.Moveable) int, resetQueueAfter bool) error
}

func ValidateMoveables(ctx context.Context, q enumerationQueue) error {
	if err := q.DequeueAndProcess(ctx, func(m *schema.Moveable) int {
		if err := validateMoveable(m); err != nil {
			slog.Warn("Skipped job: failed pre-move validation",
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.GetName(),
			)

			return queue.DecisionSkipped
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
			return queue.DecisionSkipped
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
			return queue.DecisionSkipped
		}

		return queue.DecisionSuccess
	}, true); err != nil {
		return err
	}

	return nil
}

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
