package io

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/desertwitch/gover/internal/filesystem"
)

func ensureDirectoryStructure(m *filesystem.Moveable, job *InternalProgressReport, osa osAdapter, una unixAdapter) error {
	dir := m.RootDir

	for dir != nil {
		// TO-DO: Handle generic errors here and otherwere for .Stat or .Lstat
		if _, err := osa.Stat(dir.DestPath); errors.Is(err, fs.ErrNotExist) {
			if err := una.Mkdir(dir.DestPath, dir.Metadata.Perms); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir.DestPath, err)
			}

			if err := ensurePermissions(dir.DestPath, dir.Metadata, una); err != nil {
				return fmt.Errorf("failed to ensure permissions: %w", err)
			}

			job.AnyProcessed = append(job.AnyProcessed, dir)
			job.DirsProcessed = append(job.DirsProcessed, dir)
		}
		dir = dir.Child
	}

	return nil
}
