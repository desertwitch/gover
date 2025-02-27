package io

import (
	"fmt"

	"github.com/desertwitch/gover/internal/filesystem"
)

func ensurePermissions(path string, metadata *filesystem.Metadata, una unixAdapter) error {
	if err := una.Chown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on %s: %w", path, err)
	}

	if err := una.Chmod(path, metadata.Perms); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", path, err)
	}

	return nil
}

func ensureLinkPermissions(path string, metadata *filesystem.Metadata, una unixAdapter) error {
	if err := una.Lchown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on link %s: %w", path, err)
	}

	return nil
}
