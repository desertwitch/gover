package filesystem

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

type Metadata struct {
	Inode      uint64
	Perms      uint32
	UID        uint32
	GID        uint32
	AccessedAt unix.Timespec
	ModifiedAt unix.Timespec
	Size       int64
	IsDir      bool
	IsSymlink  bool
	SymlinkTo  string
}

func getMetadata(path string) (*Metadata, error) {
	var stat unix.Stat_t

	if err := unix.Lstat(path, &stat); err != nil {
		return nil, fmt.Errorf("failed to lstat: %w", err)
	}

	metadata := &Metadata{
		Inode:      stat.Ino,
		Perms:      (uint32(stat.Mode) & 0777),
		UID:        stat.Uid,
		GID:        stat.Gid,
		AccessedAt: stat.Atim,
		ModifiedAt: stat.Mtim,
		Size:       stat.Size,
		IsDir:      (stat.Mode & unix.S_IFMT) == unix.S_IFDIR,
		IsSymlink:  (stat.Mode & unix.S_IFMT) == unix.S_IFLNK,
	}

	if metadata.IsSymlink {
		symlinkTarget, err := os.Readlink(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read symlink: %w", err)
		}
		metadata.SymlinkTo = symlinkTarget
	}

	return metadata, nil
}
