package io

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
	"golang.org/x/sys/unix"
)

type allocProvider interface{}

type fsProvider interface {
	HasEnoughFreeSpace(s unraid.Storeable, minFree uint64, fileSize uint64) (bool, error)
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

type ProgressReport struct {
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

func (i *Handler) ProcessMoveables(ctx context.Context, moveables []*filesystem.Moveable, batch *ProgressReport) error {
	for _, m := range moveables {
		if ctx.Err() != nil {
			break
		}

		job := &ProgressReport{}

		if err := i.processMoveable(ctx, m, job); err != nil {
			slog.Warn("Skipped job: failure during processing for job",
				"path", m.DestPath,
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)

			continue
		}

		slog.Info("Processed:",
			"path", m.DestPath,
			"job", m.SourcePath,
			"share", m.Share.Name)

		for _, h := range m.Hardlinks {
			if err := i.processMoveable(ctx, h, job); err != nil {
				slog.Warn("Skipped subjob: failure during processing for subjob",
					"path", h.DestPath,
					"err", err,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)

				continue
			}

			slog.Info("Processed (hardlink):",
				"path", h.DestPath,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)
		}

		for _, s := range m.Symlinks {
			if err := i.processMoveable(ctx, s, job); err != nil {
				slog.Warn("Skipped subjob: failure during processing for subjob",
					"path", s.DestPath,
					"err", err,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)

				continue
			}

			slog.Info("Processed (symlink):",
				"path", s.DestPath,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)
		}

		MergeProgressReports(batch, job)
	}

	if err := i.ensureTimestamps(batch); err != nil {
		return fmt.Errorf("failed finalizing timestamps: %w", err)
	}

	if err := i.cleanDirectoryStructure(batch); err != nil {
		return fmt.Errorf("failed cleaning source directories: %w", err)
	}

	return nil
}

func (i *Handler) processMoveable(ctx context.Context, m *filesystem.Moveable, job *ProgressReport) error {
	var jobComplete bool
	intermediateJob := &ProgressReport{}

	defer func() {
		if !jobComplete {
			i.cleanDirectoriesAfterFailure(intermediateJob) //nolint:errcheck
		}
	}()

	inUse, err := i.IsFileInUse(m.SourcePath)
	if err != nil {
		return fmt.Errorf("failed checking if source file is in use: %w", err)
	}
	if inUse {
		return ErrSourceFileInUse
	}

	if m.IsHardlink {
		if err := i.ensureDirectoryStructure(m, intermediateJob); err != nil {
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

		jobComplete = true
		MergeProgressReports(job, intermediateJob)

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.HardlinksProcessed = append(job.HardlinksProcessed, m)

		return nil
	}

	if m.IsSymlink {
		if err := i.ensureDirectoryStructure(m, intermediateJob); err != nil {
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

		jobComplete = true
		MergeProgressReports(job, intermediateJob)

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.SymlinksProcessed = append(job.SymlinksProcessed, m)

		return nil
	}

	if m.Metadata.IsSymlink {
		if err := i.ensureDirectoryStructure(m, intermediateJob); err != nil {
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

		jobComplete = true
		MergeProgressReports(job, intermediateJob)

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.MoveablesProcessed = append(job.MoveablesProcessed, m)

		return nil
	}

	if err := i.ensureDirectoryStructure(m, intermediateJob); err != nil {
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
			return ErrNotEnoughSpace
		}

		if err := i.moveFile(ctx, m); err != nil {
			return fmt.Errorf("failed to move file: %w", err)
		}
		if err := i.OSOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}
	}

	if err := i.ensurePermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("failed to ensure permissions: %w", err)
	}

	jobComplete = true
	MergeProgressReports(job, intermediateJob)

	job.AnyProcessed = append(job.AnyProcessed, m)
	job.MoveablesProcessed = append(job.MoveablesProcessed, m)

	return nil
}
