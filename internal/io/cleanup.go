package io

import (
	"log/slog"
	"sort"
)

func removeEmptyDirs(batch *InternalProgressReport, fsa fsAdapter, osa osAdapter) error {
	sort.Slice(batch.DirsProcessed, func(i, j int) bool {
		return calculateDirectoryDepth(batch.DirsProcessed[i]) > calculateDirectoryDepth(batch.DirsProcessed[j])
	})

	removed := make(map[string]struct{})

	for _, dir := range batch.DirsProcessed {
		if _, alreadyRemoved := removed[dir.SourcePath]; alreadyRemoved {
			continue
		}
		isEmpty, err := fsa.IsEmptyFolder(dir.SourcePath)
		if err != nil {
			slog.Warn("Warning (cleanup): failure establishing source directory emptiness (skipped)", "path", dir.SourcePath, "err", err)
			continue
		}
		if isEmpty {
			if err := osa.Remove(dir.SourcePath); err != nil {
				slog.Warn("Warning (cleanup): failure removing empty source directory (skipped)", "path", dir.SourcePath, "err", err)
				continue
			}
			removed[dir.SourcePath] = struct{}{}
		}
	}

	return nil
}
