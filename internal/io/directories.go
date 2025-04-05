package io

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"

	"github.com/desertwitch/gover/internal/schema"
)

// ensureDirectoryStructure recreates the necessary directory structure for a [schema.Moveable].
func (i *Handler) ensureDirectoryStructure(m *schema.Moveable, job *ioReport) error {
	dir := m.RootDir

	for dir != nil {
		if _, err := i.osHandler.Stat(dir.DestPath); errors.Is(err, fs.ErrNotExist) {
			if err := i.unixHandler.Mkdir(dir.DestPath, dir.Metadata.Perms); err != nil {
				return fmt.Errorf("(io-ensuredirs) failed to mkdir %s: %w", dir.DestPath, err)
			}

			job.AnyCreated = append(job.AnyCreated, dir)
			job.DirsCreated = append(job.DirsCreated, dir)
			job.DirsWalked = append(job.DirsWalked, dir)

			if err := i.ensurePermissions(dir.DestPath, dir.Metadata); err != nil {
				return fmt.Errorf("(io-ensuredirs) failed permissioning: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("(io-ensuredirs) failed to stat (existence): %w", err)
		} else {
			job.DirsWalked = append(job.DirsWalked, dir)
		}

		dir = dir.Child
	}

	return nil
}

// cleanDirectoryStructure deletes after moving the remaining empty directory structure.
func (i *Handler) cleanDirectoryStructure(batch *ioReport) {
	sort.Slice(batch.DirsWalked, func(i, j int) bool {
		return calculateDirectoryDepth(batch.DirsWalked[i]) > calculateDirectoryDepth(batch.DirsWalked[j])
	})

	removed := make(map[string]struct{})

	for _, dir := range batch.DirsWalked {
		if _, alreadyRemoved := removed[dir.SourcePath]; alreadyRemoved {
			continue
		}
		isEmpty, err := i.fsHandler.IsEmptyFolder(dir.SourcePath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				removed[dir.SourcePath] = struct{}{}
			} else {
				slog.Warn("Failure checking emptiness cleaning source directories (skipped)",
					"path", dir.SourcePath,
					"err", err,
				)
			}

			continue
		}
		if isEmpty {
			i.Lock()
			err := i.osHandler.Remove(dir.SourcePath)
			i.Unlock()

			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					removed[dir.SourcePath] = struct{}{}
				} else {
					slog.Warn("Failure removing directory cleaning source directories (skipped)",
						"path", dir.SourcePath,
						"err", err,
					)
				}

				continue
			}

			removed[dir.SourcePath] = struct{}{}
		}
	}
}

// cleanDirectoriesAfterFailure deletes after a failure the created empty directory structure (on target).
func (i *Handler) cleanDirectoriesAfterFailure(job *ioReport) {
	sort.Slice(job.DirsCreated, func(i, j int) bool {
		return calculateDirectoryDepth(job.DirsCreated[i]) > calculateDirectoryDepth(job.DirsCreated[j])
	})

	removed := make(map[string]struct{})

	for _, dir := range job.DirsCreated {
		if _, alreadyRemoved := removed[dir.DestPath]; alreadyRemoved {
			continue
		}
		isEmpty, err := i.fsHandler.IsEmptyFolder(dir.DestPath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				removed[dir.DestPath] = struct{}{}
			} else {
				slog.Warn("Failure checking emptiness cleaning failed directories (skipped)",
					"path", dir.DestPath,
					"err", err,
				)
			}

			continue
		}
		if isEmpty {
			i.Lock()
			err := i.osHandler.Remove(dir.DestPath)
			i.Unlock()

			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					removed[dir.DestPath] = struct{}{}
				} else {
					slog.Warn("Failure removing directory cleaning failed directories (skipped)",
						"path", dir.DestPath,
						"err", err,
					)
				}

				continue
			}

			removed[dir.DestPath] = struct{}{}
		}
	}
}

// calculateDirectoryDepth calculates a [schema.Directory] depth for use in sorting.
func calculateDirectoryDepth(dir *schema.Directory) int {
	depth := 0
	for dir != nil {
		dir = dir.Parent
		depth++
	}

	return depth
}
