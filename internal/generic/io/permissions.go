package io

import (
	"fmt"

	"github.com/desertwitch/gover/internal/generic/filesystem"
)

func (i *Handler) ensurePermissions(path string, metadata *filesystem.Metadata) error {
	if err := i.UnixOps.Chown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("(io-perms) failed to chown: %w", err)
	}

	if err := i.UnixOps.Chmod(path, metadata.Perms); err != nil {
		return fmt.Errorf("(io-perms) failed to chmod: %w", err)
	}

	return nil
}

func (i *Handler) ensureLinkPermissions(path string, metadata *filesystem.Metadata) error {
	if err := i.UnixOps.Lchown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("(io-perms) failed to lchown: %w", err)
	}

	return nil
}
