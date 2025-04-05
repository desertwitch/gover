package io

import (
	"fmt"
	"log/slog"

	"github.com/desertwitch/gover/internal/schema"
	"golang.org/x/sys/unix"
)

// ensureTimestamps recreates the timestamps on a target based on the creation
// [ioReport]. This is usually called after the entire [schema.Storage] queue is
// done processing, so that no later changes could possibly affect the
// timestamps to change from their values.
func (i *Handler) ensureTimestamps(batch *ioReport) {
	for _, a := range batch.AnyCreated {
		if err := i.ensureTimestamp(a.GetDestPath(), a.GetMetadata()); err != nil {
			slog.Warn("Failure setting a timestamp (was skipped)",
				"path", a.GetDestPath(),
				"err", err,
			)

			continue
		}
	}
}

// ensureTimestamp sets a path's timestamp based on its [schema.Metadata].
func (i *Handler) ensureTimestamp(path string, metadata *schema.Metadata) error {
	ts := []unix.Timespec{metadata.AccessedAt, metadata.ModifiedAt}
	if err := i.unixHandler.UtimesNano(path, ts); err != nil {
		return fmt.Errorf("(io-times) failed to utimesnano: %w", err)
	}

	return nil
}
