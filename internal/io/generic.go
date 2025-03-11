package io

import (
	"context"
	"errors"
	"fmt"

	"github.com/desertwitch/gover/internal/filesystem"
	"golang.org/x/sys/unix"
)

func (i *Handler) processFile(ctx context.Context, m *filesystem.Moveable) error {
	enoughSpace, err := i.FSOps.HasEnoughFreeSpace(m.Dest, m.Share.SpaceFloor, m.Metadata.Size)
	if err != nil {
		return fmt.Errorf("(io-file) failed to check enough space: %w", err)
	}
	if !enoughSpace {
		return fmt.Errorf("(io-file) %w", ErrNotEnoughSpace)
	}

	if err := i.moveFile(ctx, m); err != nil {
		return fmt.Errorf("(io-file) failed to move file: %w", err)
	}
	if err := i.OSOps.Remove(m.SourcePath); err != nil {
		return fmt.Errorf("(io-file) failed to remove src after move: %w", err)
	}
	if err := i.ensurePermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("(io-file) failed to ensure permissions: %w", err)
	}

	return nil
}

func (i *Handler) processDirectory(m *filesystem.Moveable) error {
	dirExisted := false

	if err := i.UnixOps.Mkdir(m.DestPath, m.Metadata.Perms); err != nil {
		if !errors.Is(err, unix.EEXIST) {
			return fmt.Errorf("(io-dir) failed to mkdir: %w", err)
		}
		dirExisted = true
	}

	if err := i.OSOps.Remove(m.SourcePath); err != nil {
		return fmt.Errorf("(io-dir) failed to remove src after move: %w", err)
	}

	if !dirExisted {
		if err := i.ensurePermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("(io-dir) failed to ensure permissions: %w", err)
		}
	}

	return nil
}

func (i *Handler) processHardlink(m *filesystem.Moveable) error {
	if err := i.UnixOps.Link(m.HardlinkTo.DestPath, m.DestPath); err != nil {
		return fmt.Errorf("(io-hardl) failed to link: %w", err)
	}
	if err := i.OSOps.Remove(m.SourcePath); err != nil {
		return fmt.Errorf("(io-hardl) failed to remove src after move: %w", err)
	}
	if err := i.ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("(io-hardl) failed to ensure permissions: %w", err)
	}

	return nil
}

func (i *Handler) processSymlink(m *filesystem.Moveable, internalLink bool) error {
	if internalLink {
		if err := i.UnixOps.Symlink(m.SymlinkTo.DestPath, m.DestPath); err != nil {
			return fmt.Errorf("(io-syml) failed to symlink: %w", err)
		}
	} else {
		if err := i.UnixOps.Symlink(m.Metadata.SymlinkTo, m.DestPath); err != nil {
			return fmt.Errorf("(io-syml) failed to symlink: %w", err)
		}
	}

	if err := i.OSOps.Remove(m.SourcePath); err != nil {
		return fmt.Errorf("(io-syml) failed to remove src after move: %w", err)
	}

	if err := i.ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("(io-syml) failed to ensure permissions: %w", err)
	}

	return nil
}
