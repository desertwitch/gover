package io

import (
	"fmt"
	"log/slog"

	"github.com/desertwitch/gover/internal/filesystem"
	"golang.org/x/sys/unix"
)

func (i *Handler) ensureTimestamps(batch *ProgressReport) {
	for _, a := range batch.AnyProcessed {
		if err := i.ensureTimestamp(a.GetDestPath(), a.GetMetadata()); err != nil {
			slog.Warn("Failure setting a timestamp (was skipped)",
				"path", a.GetDestPath(),
				"err", err,
			)

			continue
		}
	}
}

func (i *Handler) ensureTimestamp(path string, metadata *filesystem.Metadata) error {
	ts := []unix.Timespec{metadata.AccessedAt, metadata.ModifiedAt}
	if err := i.UnixOps.UtimesNano(path, ts); err != nil {
		return fmt.Errorf("(io-times) failed to utimesnano: %w", err)
	}

	return nil
}
