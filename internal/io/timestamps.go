package io

import (
	"fmt"
	"log/slog"

	"github.com/desertwitch/gover/internal/filesystem"
	"golang.org/x/sys/unix"
)

func (i *IOHandler) ensureTimestamps(batch *InternalProgressReport) error {
	for _, a := range batch.AnyProcessed {
		if err := i.ensureTimestamp(a.GetDestPath(), a.GetMetadata()); err != nil {
			slog.Warn("Warning (finalize): failure setting timestamp", "path", a.GetDestPath(), "err", err)
			continue
		}
	}
	return nil
}

func (i *IOHandler) ensureTimestamp(path string, metadata *filesystem.Metadata) error {
	ts := []unix.Timespec{metadata.AccessedAt, metadata.ModifiedAt}
	if err := i.UnixOps.UtimesNano(path, ts); err != nil {
		return fmt.Errorf("failed to set timestamp: %w", err)
	}
	return nil
}
