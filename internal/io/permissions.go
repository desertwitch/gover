package io

import (
	"fmt"

	"github.com/desertwitch/gover/internal/schema"
)

// ensurePermissions sets permissions and ownership for a given path based on its [schema.Metadata].
func (i *Handler) ensurePermissions(path string, metadata *schema.Metadata) error {
	if err := i.unixHandler.Chown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("(io-perms) failed to chown: %w", err)
	}

	if err := i.unixHandler.Chmod(path, metadata.Perms); err != nil {
		return fmt.Errorf("(io-perms) failed to chmod: %w", err)
	}

	return nil
}

// ensureLinkPermissions sets permissions and ownership for a given link path based on its [schema.Metadata].
func (i *Handler) ensureLinkPermissions(path string, metadata *schema.Metadata) error {
	if err := i.unixHandler.Lchown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("(io-perms) failed to lchown: %w", err)
	}

	return nil
}
