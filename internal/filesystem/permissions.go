package filesystem

import (
	"fmt"
)

func (f *FileHandler) ensurePermissions(path string, metadata *Metadata) error {
	if err := f.UnixOps.Chown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on %s: %w", path, err)
	}

	if err := f.UnixOps.Chmod(path, metadata.Perms); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", path, err)
	}

	return nil
}

func (f *FileHandler) ensureLinkPermissions(path string, metadata *Metadata) error {
	if err := f.UnixOps.Lchown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on link %s: %w", path, err)
	}

	return nil
}
