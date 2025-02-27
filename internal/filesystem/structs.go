package filesystem

import (
	"github.com/desertwitch/gover/internal/unraid"
	"golang.org/x/sys/unix"
)

type RelatedElement interface {
	GetMetadata() *Metadata
	GetSourcePath() string
	GetDestPath() string
}

type Moveable struct {
	Share      *unraid.UnraidShare
	Source     unraid.UnraidStoreable
	SourcePath string
	Dest       unraid.UnraidStoreable
	DestPath   string
	Hardlinks  []*Moveable
	Hardlink   bool
	HardlinkTo *Moveable
	Symlinks   []*Moveable
	Symlink    bool
	SymlinkTo  *Moveable
	Metadata   *Metadata
	RootDir    *RelatedDirectory
	DeepestDir *RelatedDirectory
}

func (m *Moveable) GetMetadata() *Metadata {
	return m.Metadata
}

func (m *Moveable) GetSourcePath() string {
	return m.SourcePath
}

func (m *Moveable) GetDestPath() string {
	return m.DestPath
}

type RelatedDirectory struct {
	SourcePath string
	DestPath   string
	Metadata   *Metadata
	Parent     *RelatedDirectory
	Child      *RelatedDirectory
}

func (d *RelatedDirectory) GetMetadata() *Metadata {
	return d.Metadata
}

func (d *RelatedDirectory) GetSourcePath() string {
	return d.SourcePath
}

func (d *RelatedDirectory) GetDestPath() string {
	return d.DestPath
}

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

type DiskStats struct {
	TotalSize int64
	FreeSpace int64
}
