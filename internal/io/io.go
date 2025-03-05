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
	IsFileInUse(path string) (bool, error)
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

		mergeProgressReports(batch, job)
	}

	i.ensureTimestamps(batch)
	i.cleanDirectoryStructure(batch)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	return nil
}

func (i *Handler) processMoveable(ctx context.Context, m *filesystem.Moveable, job *ProgressReport) error {
	var jobComplete bool

	intermediateJob := &ProgressReport{}

	defer func() {
		if jobComplete {
			addToProgressReport(intermediateJob, m)
			mergeProgressReports(job, intermediateJob)
		} else {
			i.cleanDirectoriesAfterFailure(intermediateJob)
		}
	}()

	if inUse, err := i.FSOps.IsFileInUse(m.SourcePath); err != nil {
		return fmt.Errorf("failed checking if source file is in use: %w", err)
	} else if inUse {
		return ErrSourceFileInUse
	}

	if err := i.ensureDirectoryStructure(m, intermediateJob); err != nil {
		return fmt.Errorf("failed to ensure dir structure: %w", err)
	}

	if !m.Metadata.IsDir && !m.IsHardlink && !m.IsSymlink && !m.Metadata.IsSymlink {
		if err := i.processFile(ctx, m); err != nil {
			return fmt.Errorf("failed to process file: %w", err)
		}
		jobComplete = true
	}

	if m.Metadata.IsDir {
		if err := i.processDirectory(m); err != nil {
			return fmt.Errorf("failed to process directory: %w", err)
		}
		jobComplete = true
	}

	if m.IsHardlink {
		if err := i.processHardlink(m); err != nil {
			return fmt.Errorf("failed to process hardlink: %w", err)
		}
		jobComplete = true
	}

	if m.IsSymlink {
		if err := i.processSymlink(m, true); err != nil {
			return fmt.Errorf("failed to process symlink: %w", err)
		}
		jobComplete = true
	}

	if m.Metadata.IsSymlink {
		if err := i.processSymlink(m, false); err != nil {
			return fmt.Errorf("failed to process symlink: %w", err)
		}
		jobComplete = true
	}

	if !jobComplete {
		return ErrNothingToProcess
	}

	return nil
}
