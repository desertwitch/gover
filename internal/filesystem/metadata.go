package filesystem

import (
	"fmt"

	"golang.org/x/sys/unix"
)

type Metadata struct {
	Inode      uint64
	Perms      uint32
	UID        uint32
	GID        uint32
	AccessedAt unix.Timespec
	ModifiedAt unix.Timespec
	Size       uint64
	IsDir      bool
	IsSymlink  bool
	SymlinkTo  string
}

func (f *Handler) getMetadata(path string) (*Metadata, error) {
	var stat unix.Stat_t

	if err := f.UnixOps.Lstat(path, &stat); err != nil {
		return nil, fmt.Errorf("(fs-metadata) failed to lstat: %w", err)
	}

	metadata := &Metadata{
		Inode:      stat.Ino,
		Perms:      stat.Mode & 0o777, //nolint:mnd
		UID:        stat.Uid,
		GID:        stat.Gid,
		AccessedAt: stat.Atim,
		ModifiedAt: stat.Mtim,
		Size:       handleSize(stat.Size),
		IsDir:      (stat.Mode & unix.S_IFMT) == unix.S_IFDIR,
		IsSymlink:  (stat.Mode & unix.S_IFMT) == unix.S_IFLNK,
	}

	if metadata.IsSymlink {
		symlinkTarget, err := f.OSOps.Readlink(path)
		if err != nil {
			return nil, fmt.Errorf("(fs-metadata) failed to read symlink: %w", err)
		}
		metadata.SymlinkTo = symlinkTarget
	}

	return metadata, nil
}
