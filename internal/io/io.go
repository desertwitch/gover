package io

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
	"github.com/zeebo/blake3"
	"golang.org/x/sys/unix"
)

type allocProvider interface{}

type fsProvider interface {
	HasEnoughFreeSpace(s unraid.Storeable, minFree int64, fileSize int64) (bool, error)
	IsEmptyFolder(path string) (bool, error)
}

type osProvider interface {
	Open(name string) (*os.File, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Remove(name string) error
	Rename(oldpath, newpath string) error
	Stat(name string) (os.FileInfo, error)
}

type unixProvider interface {
	Chmod(path string, mode uint32) error
	Chown(path string, uid, gid int) error
	Lchown(path string, uid, gid int) error
	Link(oldpath, newpath string) error
	Mkdir(path string, mode uint32) error
	Symlink(oldpath, newpath string) error
	UtimesNano(path string, times []unix.Timespec) error
}

type relatedElement interface {
	GetDestPath() string
	GetMetadata() *filesystem.Metadata
	GetSourcePath() string
}

type InternalProgressReport struct {
	AnyProcessed       []relatedElement
	DirsProcessed      []*filesystem.RelatedDirectory
	MoveablesProcessed []*filesystem.Moveable
	SymlinksProcessed  []*filesystem.Moveable
	HardlinksProcessed []*filesystem.Moveable
}

type Handler struct {
	AllocOps allocProvider
	FSOps    fsProvider
	OSOps    osProvider
	UnixOps  unixProvider
}

func NewHandler(allocOps allocProvider, fsOps fsProvider, osOps osProvider, unixOps unixProvider) *Handler {
	return &Handler{
		AllocOps: allocOps,
		FSOps:    fsOps,
		OSOps:    osOps,
		UnixOps:  unixOps,
	}
}

// TO-DO:
// Reallocation if not enough space (up to 3x?)
// Rollback, Locking?

func (i *Handler) ProcessMoveables(moveables []*filesystem.Moveable, batch *InternalProgressReport) error {
	for _, m := range moveables {
		job := &InternalProgressReport{}

		if err := i.processMoveable(m, job); err != nil {
			slog.Warn("Skipped job: failure during processing for job", "path", m.DestPath, "err", err, "job", m.SourcePath, "share", m.Share.Name)

			continue
		}
		slog.Info("Processed:", "path", m.DestPath, "job", m.SourcePath, "share", m.Share.Name)

		for _, h := range m.Hardlinks {
			if err := i.processMoveable(h, job); err != nil {
				slog.Warn("Skipped subjob: failure during processing for subjob", "path", h.DestPath, "err", err, "job", m.SourcePath, "share", m.Share.Name)

				continue
			}
			slog.Info("Processed (hardlink):", "path", h.DestPath, "job", m.SourcePath, "share", m.Share.Name)
		}

		for _, s := range m.Symlinks {
			if err := i.processMoveable(s, job); err != nil {
				slog.Warn("Skipped subjob: failure during processing for subjob", "path", s.DestPath, "err", err, "job", m.SourcePath, "share", m.Share.Name)

				continue
			}
			slog.Info("Processed (symlink):", "path", s.DestPath, "job", m.SourcePath, "share", m.Share.Name)
		}

		batch.AnyProcessed = append(batch.AnyProcessed, job.AnyProcessed...)
		batch.DirsProcessed = append(batch.DirsProcessed, job.DirsProcessed...)
		batch.HardlinksProcessed = append(batch.HardlinksProcessed, job.HardlinksProcessed...)
		batch.MoveablesProcessed = append(batch.MoveablesProcessed, job.MoveablesProcessed...)
		batch.SymlinksProcessed = append(batch.SymlinksProcessed, job.SymlinksProcessed...)
	}

	if err := i.ensureTimestamps(batch); err != nil {
		return fmt.Errorf("failed finalizing timestamps: %w", err)
	}

	if err := i.cleanDirectoryStructure(batch); err != nil {
		return fmt.Errorf("failed cleaning source directories: %w", err)
	}

	return nil
}

func (i *Handler) processMoveable(m *filesystem.Moveable, job *InternalProgressReport) error {
	used, err := i.IsFileInUse(m.SourcePath)
	if err != nil {
		return fmt.Errorf("failed checking if source file is in use: %w", err)
	}
	if used {
		return errors.New("source file is currently in use")
	}

	if m.Hardlink {
		if err := i.ensureDirectoryStructure(m, job); err != nil {
			return fmt.Errorf("failed to ensure dir tree for hardlink: %w", err)
		}

		if err := i.UnixOps.Link(m.HardlinkTo.DestPath, m.DestPath); err != nil {
			return fmt.Errorf("failed to create hardlink: %w", err)
		}
		if err := i.OSOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}

		if err := i.ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.HardlinksProcessed = append(job.HardlinksProcessed, m)

		return nil
	}

	if m.Symlink {
		if err := i.ensureDirectoryStructure(m, job); err != nil {
			return fmt.Errorf("failed to ensure dir tree for symlink: %w", err)
		}

		if err := i.UnixOps.Symlink(m.SymlinkTo.DestPath, m.DestPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
		if err := i.OSOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}

		if err := i.ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.SymlinksProcessed = append(job.SymlinksProcessed, m)

		return nil
	}

	if m.Metadata.IsSymlink {
		if err := i.ensureDirectoryStructure(m, job); err != nil {
			return fmt.Errorf("failed to ensure dir tree: %w", err)
		}

		if err := i.UnixOps.Symlink(m.Metadata.SymlinkTo, m.DestPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
		if err := i.OSOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}

		if err := i.ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.MoveablesProcessed = append(job.MoveablesProcessed, m)

		return nil
	}

	if err := i.ensureDirectoryStructure(m, job); err != nil {
		return fmt.Errorf("failed to ensure dir tree: %w", err)
	}

	if m.Metadata.IsDir {
		if err := i.UnixOps.Mkdir(m.DestPath, m.Metadata.Perms); err != nil {
			return fmt.Errorf("failed to create empty dir: %w", err)
		}
		if err := i.OSOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}
	} else {
		enoughSpace, err := i.FSOps.HasEnoughFreeSpace(m.Dest, m.Share.SpaceFloor, m.Metadata.Size)
		if err != nil {
			return fmt.Errorf("failed to check for enough space: %w", err)
		}
		if !enoughSpace {
			if _, ok := m.Dest.(*unraid.Disk); ok {
				// TO-DO: Reallocate with hardlinks
			} else {
				return errors.New("not enough free space on destination pool")
			}
		}

		if err := i.moveFile(m); err != nil {
			return fmt.Errorf("failed to move file: %w", err)
		}
		if err := i.OSOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}
	}

	if err := i.ensurePermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("failed to ensure permissions: %w", err)
	}

	job.AnyProcessed = append(job.AnyProcessed, m)
	job.MoveablesProcessed = append(job.MoveablesProcessed, m)

	return nil
}

func (i *Handler) moveFile(m *filesystem.Moveable) error {
	srcFile, err := i.OSOps.Open(m.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	tmpPath := m.DestPath + ".gover"
	defer func() {
		if err != nil {
			i.OSOps.Remove(tmpPath)
		}
	}()

	dstFile, err := i.OSOps.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, os.FileMode(m.Metadata.Perms))
	if err != nil {
		return fmt.Errorf("failed to open destination file %s: %w", tmpPath, err)
	}
	defer dstFile.Close()

	srcHasher := blake3.New()
	dstHasher := blake3.New()

	teeReader := io.TeeReader(srcFile, srcHasher)
	multiWriter := io.MultiWriter(dstFile, dstHasher)

	if _, err := io.Copy(multiWriter, teeReader); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination fs: %w", err)
	}

	srcChecksum := hex.EncodeToString(srcHasher.Sum(nil))
	dstChecksum := hex.EncodeToString(dstHasher.Sum(nil))

	if srcChecksum != dstChecksum {
		return fmt.Errorf("hash mismatch: %s (src) != %s (dst)", srcChecksum, dstChecksum)
	}

	if _, err := i.OSOps.Stat(m.DestPath); err == nil {
		return errors.New("rename destination already exists")
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to check rename destination existence: %w", err)
	}

	if err := i.OSOps.Rename(tmpPath, m.DestPath); err != nil {
		return fmt.Errorf("failed to rename temporary file to destination file: %w", err)
	}

	return nil
}
