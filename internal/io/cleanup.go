package io

import (
	"log/slog"
	"os"
	"sort"

	"github.com/desertwitch/gover/internal/filesystem"
)

func removeEmptyDirs(batch *InternalProgressReport) error {
	sort.Slice(batch.DirsProcessed, func(i, j int) bool {
		return calculateDirectoryDepth(batch.DirsProcessed[i]) > calculateDirectoryDepth(batch.DirsProcessed[j])
	})

	removed := make(map[string]struct{})

	for _, dir := range batch.DirsProcessed {
		if _, alreadyRemoved := removed[dir.SourcePath]; alreadyRemoved {
			continue
		}
		isEmpty, err := filesystem.IsEmptyFolder(dir.SourcePath)
		if err != nil {
			slog.Warn("Warning (cleanup): failure establishing source directory emptiness (skipped)", "path", dir.SourcePath, "err", err)
			continue
		}
		if isEmpty {
			if err := os.Remove(dir.SourcePath); err != nil {
				slog.Warn("Warning (cleanup): failure removing empty source directory (skipped)", "path", dir.SourcePath, "err", err)
				continue
			}
			removed[dir.SourcePath] = struct{}{}
		}
	}

	return nil
}
