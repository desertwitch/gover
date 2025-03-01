package filesystem

import (
	"fmt"
	"log/slog"

	"golang.org/x/sys/unix"
)

func (f *FileHandler) ensureTimestamps(batch *InternalProgressReport) error {
	for _, a := range batch.AnyProcessed {
		if err := f.ensureTimestamp(a.GetDestPath(), a.GetMetadata()); err != nil {
			slog.Warn("Warning (finalize): failure setting timestamp", "path", a.GetDestPath(), "err", err)
			continue
		}
	}
	return nil
}

func (f *FileHandler) ensureTimestamp(path string, metadata *Metadata) error {
	ts := []unix.Timespec{metadata.AccessedAt, metadata.ModifiedAt}
	if err := f.UnixOps.UtimesNano(path, ts); err != nil {
		return fmt.Errorf("failed to set timestamp: %w", err)
	}
	return nil
}
