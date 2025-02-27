package io

import (
	"fmt"

	"github.com/desertwitch/gover/internal/filesystem"
	"golang.org/x/sys/unix"
)

func ensurePermissions(path string, metadata *filesystem.Metadata) error {
	if err := unix.Chown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on %s: %w", path, err)
	}

	if err := unix.Chmod(path, metadata.Perms); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", path, err)
	}

	return nil
}

func ensureLinkPermissions(path string, metadata *filesystem.Metadata) error {
	if err := unix.Lchown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on link %s: %w", path, err)
	}

	return nil
}
