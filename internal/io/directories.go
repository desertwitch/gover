package io

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"

	"github.com/desertwitch/gover/internal/filesystem"
)

func (i *Handler) ensureDirectoryStructure(m *filesystem.Moveable, job *ProgressReport) error {
	dir := m.RootDir

	for dir != nil {
		if _, err := i.OSOps.Stat(dir.DestPath); errors.Is(err, fs.ErrNotExist) {
			if err := i.UnixOps.Mkdir(dir.DestPath, dir.Metadata.Perms); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir.DestPath, err)
			}

			job.AnyProcessed = append(job.AnyProcessed, dir)
			job.DirsProcessed = append(job.DirsProcessed, dir)

			if err := i.ensurePermissions(dir.DestPath, dir.Metadata); err != nil {
				return fmt.Errorf("failed to ensure permissions: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed checking folder while ensuring dir structure: %w", err)
		}
		dir = dir.Child
	}

	return nil
}

func (i *Handler) cleanDirectoryStructure(batch *ProgressReport) error {
	sort.Slice(batch.DirsProcessed, func(i, j int) bool {
		return calculateDirectoryDepth(batch.DirsProcessed[i]) > calculateDirectoryDepth(batch.DirsProcessed[j])
	})

	removed := make(map[string]struct{})

	for _, dir := range batch.DirsProcessed {
		if _, alreadyRemoved := removed[dir.SourcePath]; alreadyRemoved {
			continue
		}
		isEmpty, err := i.FSOps.IsEmptyFolder(dir.SourcePath)
		if err != nil {
			slog.Warn("Warning (cleanup): failure establishing source directory emptiness (skipped)",
				"path", dir.SourcePath,
				"err", err,
			)

			continue
		}
		if isEmpty {
			if err := i.OSOps.Remove(dir.SourcePath); err != nil {
				slog.Warn("Warning (cleanup): failure removing empty source directory (skipped)",
					"path", dir.SourcePath,
					"err", err,
				)

				continue
			}
			removed[dir.SourcePath] = struct{}{}
		}
	}

	return nil
}

func (i *Handler) cleanDirectoriesAfterFailure(job *ProgressReport) error {
	sort.Slice(job.DirsProcessed, func(i, j int) bool {
		return calculateDirectoryDepth(job.DirsProcessed[i]) > calculateDirectoryDepth(job.DirsProcessed[j])
	})

	removed := make(map[string]struct{})

	for _, dir := range job.DirsProcessed {
		if _, alreadyRemoved := removed[dir.DestPath]; alreadyRemoved {
			continue
		}
		isEmpty, err := i.FSOps.IsEmptyFolder(dir.DestPath)
		if err != nil {
			slog.Warn("Warning (cleanup): failure establishing dest directory emptiness (skipped)",
				"path", dir.DestPath,
				"err", err,
			)

			continue
		}
		if isEmpty {
			if err := i.OSOps.Remove(dir.DestPath); err != nil {
				slog.Warn("Warning (cleanup): failure removing empty dest directory (skipped)",
					"path", dir.DestPath,
					"err", err,
				)

				continue
			}
			removed[dir.DestPath] = struct{}{}
		}
	}

	return nil
}

func calculateDirectoryDepth(dir *filesystem.RelatedDirectory) int {
	depth := 0
	for dir != nil {
		dir = dir.Parent
		depth++
	}

	return depth
}
