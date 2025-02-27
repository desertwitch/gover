package io

import (
	"fmt"

	"github.com/desertwitch/gover/internal/filesystem"
)

func ensurePermissions(path string, metadata *filesystem.Metadata, unixOps unixProvider) error {
	if err := unixOps.Chown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on %s: %w", path, err)
	}

	if err := unixOps.Chmod(path, metadata.Perms); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", path, err)
	}

	return nil
}

func ensureLinkPermissions(path string, metadata *filesystem.Metadata, unixOps unixProvider) error {
	if err := unixOps.Lchown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on link %s: %w", path, err)
	}

	return nil
}
