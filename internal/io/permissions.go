package io

import (
	"fmt"

	"github.com/desertwitch/gover/internal/filesystem"
)

func (i *Handler) ensurePermissions(path string, metadata *filesystem.Metadata) error {
	if err := i.UnixOps.Chown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on %s: %w", path, err)
	}

	if err := i.UnixOps.Chmod(path, metadata.Perms); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", path, err)
	}

	return nil
}

func (i *Handler) ensureLinkPermissions(path string, metadata *filesystem.Metadata) error {
	if err := i.UnixOps.Lchown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on link %s: %w", path, err)
	}

	return nil
}
