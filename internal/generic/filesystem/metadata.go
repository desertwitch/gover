package filesystem

import (
	"fmt"
	"log/slog"

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

const (
	unixBasePerms = 0o777
)

func (f *Handler) establishMetadata(m *Moveable) error {
	metadata, err := f.getMetadata(m.SourcePath)
	if err != nil {
		slog.Warn("Skipped job: failed to get metadata",
			"err", err,
			"job", m.SourcePath,
			"share", m.Share.GetName(),
		)

		return err
	}
	m.Metadata = metadata

	return nil
}

func (f *Handler) getMetadata(path string) (*Metadata, error) {
	var stat unix.Stat_t

	if err := f.unixHandler.Lstat(path, &stat); err != nil {
		return nil, fmt.Errorf("(fs-metadata) failed to lstat: %w", err)
	}

	metadata := &Metadata{
		Inode:      stat.Ino,
		Perms:      stat.Mode & unixBasePerms,
		UID:        stat.Uid,
		GID:        stat.Gid,
		AccessedAt: stat.Atim,
		ModifiedAt: stat.Mtim,
		Size:       handleSize(stat.Size),
		IsDir:      (stat.Mode & unix.S_IFMT) == unix.S_IFDIR,
		IsSymlink:  (stat.Mode & unix.S_IFMT) == unix.S_IFLNK,
	}

	if metadata.IsSymlink {
		symlinkTarget, err := f.osHandler.Readlink(path)
		if err != nil {
			return nil, fmt.Errorf("(fs-metadata) failed to readlink: %w", err)
		}
		metadata.SymlinkTo = symlinkTarget
	}

	return metadata, nil
}
