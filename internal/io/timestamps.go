package io

import (
	"fmt"
	"log/slog"

	"github.com/desertwitch/gover/internal/filesystem"
	"golang.org/x/sys/unix"
)

func ensureTimestamps(batch *InternalProgressReport, una unixAdapter) error {
	for _, a := range batch.AnyProcessed {
		if err := ensureTimestamp(a.GetDestPath(), a.GetMetadata(), una); err != nil {
			slog.Warn("Warning (finalize): failure setting timestamp", "path", a.GetDestPath(), "err", err)
			continue
		}
	}
	return nil
}

func ensureTimestamp(path string, metadata *filesystem.Metadata, una unixAdapter) error {
	ts := []unix.Timespec{metadata.AccessedAt, metadata.ModifiedAt}
	if err := una.UtimesNano(path, ts); err != nil {
		return fmt.Errorf("failed to set timestamp: %w", err)
	}
	return nil
}
