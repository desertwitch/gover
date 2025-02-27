package io

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/desertwitch/gover/internal/filesystem"
	"golang.org/x/sys/unix"
)

func ensureDirectoryStructure(m *filesystem.Moveable, job *InternalProgressReport) error {
	dir := m.RootDir

	for dir != nil {
		// TO-DO: Handle generic errors here and otherwere for .Stat or .Lstat
		if _, err := os.Stat(dir.DestPath); errors.Is(err, fs.ErrNotExist) {
			if err := unix.Mkdir(dir.DestPath, dir.Metadata.Perms); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir.DestPath, err)
			}

			if err := ensurePermissions(dir.DestPath, dir.Metadata); err != nil {
				return fmt.Errorf("failed to ensure permissions: %w", err)
			}

			job.AnyProcessed = append(job.AnyProcessed, dir)
			job.DirsProcessed = append(job.DirsProcessed, dir)
		}
		dir = dir.Child
	}

	return nil
}
