package schema

import "golang.org/x/sys/unix"

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
