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
				return fmt.Errorf("(io-ensuredirs) failed to mkdir %s: %w", dir.DestPath, err)
			}

			job.AnyCreated = append(job.AnyCreated, dir)
			job.DirsCreated = append(job.DirsCreated, dir)

			if err := i.ensurePermissions(dir.DestPath, dir.Metadata); err != nil {
				return fmt.Errorf("(io-ensuredirs) failed permissioning: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("(io-ensuredirs) failed to stat (existence): %w", err)
		} else {
			job.DirsProcessed = append(job.DirsProcessed, dir)
		}

		dir = dir.Child
	}

	return nil
}

func (i *Handler) cleanDirectoryStructure(batch *ProgressReport) {
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
			if !errors.Is(err, fs.ErrNotExist) {
				slog.Warn("Failure checking emptiness cleaning source directories (skipped)",
					"path", dir.SourcePath,
					"err", err,
				)
			} else {
				removed[dir.SourcePath] = struct{}{}
			}

			continue
		}
		if isEmpty {
			i.Lock()
			err := i.OSOps.Remove(dir.SourcePath)
			i.Unlock()

			if err != nil {
				slog.Warn("Failure removing directory cleaning source directories (skipped)",
					"path", dir.SourcePath,
					"err", err,
				)

				continue
			}

			removed[dir.SourcePath] = struct{}{}
		}
	}
}

func (i *Handler) cleanDirectoriesAfterFailure(job *ProgressReport) {
	sort.Slice(job.DirsCreated, func(i, j int) bool {
		return calculateDirectoryDepth(job.DirsCreated[i]) > calculateDirectoryDepth(job.DirsCreated[j])
	})

	removed := make(map[string]struct{})

	for _, dir := range job.DirsCreated {
		if _, alreadyRemoved := removed[dir.DestPath]; alreadyRemoved {
			continue
		}
		isEmpty, err := i.FSOps.IsEmptyFolder(dir.DestPath)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				slog.Warn("Failure checking emptiness cleaning failed directories (skipped)",
					"path", dir.DestPath,
					"err", err,
				)
			} else {
				removed[dir.DestPath] = struct{}{}
			}

			continue
		}
		if isEmpty {
			if err := i.OSOps.Remove(dir.DestPath); err != nil {
				slog.Warn("Failure removing directory cleaning failed directories (skipped)",
					"path", dir.DestPath,
					"err", err,
				)

				continue
			}
			removed[dir.DestPath] = struct{}{}
		}
	}
}

func calculateDirectoryDepth(dir *filesystem.RelatedDirectory) int {
	depth := 0
	for dir != nil {
		dir = dir.Parent
		depth++
	}

	return depth
}
