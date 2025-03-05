package io

import (
	"context"
	"fmt"

	"github.com/desertwitch/gover/internal/filesystem"
)

func (i *Handler) processFile(ctx context.Context, m *filesystem.Moveable) error {
	enoughSpace, err := i.FSOps.HasEnoughFreeSpace(m.Dest, m.Share.SpaceFloor, m.Metadata.Size)
	if err != nil {
		return fmt.Errorf("failed to check for enough space: %w", err)
	}
	if !enoughSpace {
		return ErrNotEnoughSpace
	}

	if err := i.moveFile(ctx, m); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}
	if err := i.OSOps.Remove(m.SourcePath); err != nil {
		return fmt.Errorf("failed to remove source after move: %w", err)
	}
	if err := i.ensurePermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("failed to ensure permissions: %w", err)
	}

	return nil
}

func (i *Handler) processDirectory(m *filesystem.Moveable) error {
	if err := i.UnixOps.Mkdir(m.DestPath, m.Metadata.Perms); err != nil {
		return fmt.Errorf("failed to create empty dir: %w", err)
	}
	if err := i.OSOps.Remove(m.SourcePath); err != nil {
		return fmt.Errorf("failed to remove source after move: %w", err)
	}
	if err := i.ensurePermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("failed to ensure permissions: %w", err)
	}

	return nil
}

func (i *Handler) processHardlink(m *filesystem.Moveable) error {
	if err := i.UnixOps.Link(m.HardlinkTo.DestPath, m.DestPath); err != nil {
		return fmt.Errorf("failed to create hardlink: %w", err)
	}
	if err := i.OSOps.Remove(m.SourcePath); err != nil {
		return fmt.Errorf("failed to remove source after move: %w", err)
	}
	if err := i.ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("failed to ensure link permissions: %w", err)
	}

	return nil
}

func (i *Handler) processSymlink(m *filesystem.Moveable, internalLink bool) error {
	if internalLink {
		if err := i.UnixOps.Symlink(m.SymlinkTo.DestPath, m.DestPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
	} else {
		if err := i.UnixOps.Symlink(m.Metadata.SymlinkTo, m.DestPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
	}

	if err := i.OSOps.Remove(m.SourcePath); err != nil {
		return fmt.Errorf("failed to remove source after move: %w", err)
	}

	if err := i.ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("failed to ensure link permissions: %w", err)
	}

	return nil
}
