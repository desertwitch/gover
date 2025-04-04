package io

import (
	"fmt"
	"log/slog"

	"github.com/desertwitch/gover/internal/schema"
	"golang.org/x/sys/unix"
)

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

func (i *Handler) ensureTimestamp(path string, metadata *schema.Metadata) error {
	ts := []unix.Timespec{metadata.AccessedAt, metadata.ModifiedAt}
	if err := i.unixHandler.UtimesNano(path, ts); err != nil {
		return fmt.Errorf("(io-times) failed to utimesnano: %w", err)
	}

	return nil
}
