package io

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/desertwitch/gover/internal/schema"
	"golang.org/x/sys/unix"
)

// cleanFileAfterFailure removes after a failure the destination file, provided
// that the respective source file still exists.
func (i *Handler) cleanFileAfterFailure(m *schema.Moveable) {
	if _, err := i.osHandler.Stat(m.SourcePath); err == nil {
		if err := i.osHandler.Remove(m.DestPath); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				slog.Warn("Failure removing destination file cleaning after failure (skipped)",
					"path", m.DestPath,
					"err", err,
				)
			}
		}
	}
}

// processFile is the principal method for IO-processing a file-type
// [schema.Moveable]. Apart from moving the file itself, it handles both
// spacing, permissioning and cleanup.
func (i *Handler) processFile(ctx context.Context, m *schema.Moveable) error {
	enoughSpace, err := i.fsHandler.HasEnoughFreeSpace(m.Dest, m.Share.GetSpaceFloor(), m.Metadata.Size)
	if err != nil {
		return fmt.Errorf("(io-file) failed to check enough space: %w", err)
	}
	if !enoughSpace {
		return fmt.Errorf("(io-file) %w", ErrNotEnoughSpace)
	}

	if err := i.moveFile(ctx, m); err != nil {
		return fmt.Errorf("(io-file) failed to move file: %w", err)
	}

	if err := i.ensurePermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("(io-file) failed to ensure permissions: %w", err)
	}

	if err := i.osHandler.Remove(m.SourcePath); err != nil {
		return fmt.Errorf("(io-file) failed to remove src after move: %w", err)
	}

	return nil
}

// processDirectory is the principal method for IO-processing a directory-type
// [schema.Moveable]. Apart from recreating the directory itself, it handles
// both permissioning and cleanup as well.
func (i *Handler) processDirectory(m *schema.Moveable) error {
	dirExisted := false

	if err := i.unixHandler.Mkdir(m.DestPath, m.Metadata.Perms); err != nil {
		if !errors.Is(err, unix.EEXIST) {
			return fmt.Errorf("(io-dir) failed to mkdir: %w", err)
		}
		dirExisted = true
	}

	if !dirExisted {
		if err := i.ensurePermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("(io-dir) failed to ensure permissions: %w", err)
		}
	}

	if err := i.osHandler.Remove(m.SourcePath); err != nil {
		return fmt.Errorf("(io-dir) failed to remove src after move: %w", err)
	}

	return nil
}

// processHardlink is the principal method for IO-processing a hardlink-type
// [schema.Moveable]. Apart from recreating the hardlink itself, it handles both
// permissioning and cleanup as well.
func (i *Handler) processHardlink(m *schema.Moveable) error {
	if err := i.unixHandler.Link(m.HardlinkTo.DestPath, m.DestPath); err != nil {
		return fmt.Errorf("(io-hardl) failed to link: %w", err)
	}

	if err := i.ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("(io-hardl) failed to ensure permissions: %w", err)
	}

	if err := i.osHandler.Remove(m.SourcePath); err != nil {
		return fmt.Errorf("(io-hardl) failed to remove src after move: %w", err)
	}

	return nil
}

// processSymlink is the principal method for IO-processing a symlink-type
// [schema.Moveable]. Apart from recreating the symlink itself, it handles both
// permissioning and cleanup as well.
func (i *Handler) processSymlink(m *schema.Moveable, internalLink bool) error {
	if internalLink {
		if err := i.unixHandler.Symlink(m.SymlinkTo.DestPath, m.DestPath); err != nil {
			return fmt.Errorf("(io-syml) failed to symlink: %w", err)
		}
	} else {
		if err := i.unixHandler.Symlink(m.Metadata.SymlinkTo, m.DestPath); err != nil {
			return fmt.Errorf("(io-syml) failed to symlink: %w", err)
		}
	}

	if err := i.ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("(io-syml) failed to ensure permissions: %w", err)
	}

	if err := i.osHandler.Remove(m.SourcePath); err != nil {
		return fmt.Errorf("(io-syml) failed to remove src after move: %w", err)
	}

	return nil
}
