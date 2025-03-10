package io

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/queue"
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

func (i *Handler) ProcessQueue(ctx context.Context, q *queue.QueueManager) error {
	batch := &ProgressReport{}

	for {
		if ctx.Err() != nil {
			break
		}

		m, ok := q.Dequeue()
		if !ok {
			break
		}

		job := &ProgressReport{}

		q.SetProcessing(m)
		if err := i.processMoveable(ctx, m, job); err != nil {
			slog.Warn("Skipped job: failure during processing",
				"path", m.DestPath,
				"err", err,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)
			q.SetSkipped(m)

			continue
		}
		q.SetSuccess(m)

		slog.Info("Processed:",
			"path", m.DestPath,
			"job", m.SourcePath,
			"share", m.Share.Name)

		for _, h := range m.Hardlinks {
			q.SetProcessing(h)
			if err := i.processMoveable(ctx, h, job); err != nil {
				slog.Warn("Skipped subjob: failure during processing",
					"path", h.DestPath,
					"err", err,
					"subjob", h.SourcePath,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)
				q.SetSkipped(h)

				continue
			}
			q.SetSuccess(h)

			slog.Info("Processed (hardlink):",
				"path", h.DestPath,
				"subjob", h.SourcePath,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)
		}

		for _, s := range m.Symlinks {
			q.SetProcessing(s)
			if err := i.processMoveable(ctx, s, job); err != nil {
				slog.Warn("Skipped subjob: failure during processing",
					"path", s.DestPath,
					"err", err,
					"subjob", s.SourcePath,
					"job", m.SourcePath,
					"share", m.Share.Name,
				)
				q.SetSkipped(s)

				continue
			}
			q.SetSuccess(s)

			slog.Info("Processed (symlink):",
				"path", s.DestPath,
				"subjob", s.SourcePath,
				"job", m.SourcePath,
				"share", m.Share.Name,
			)
		}

		mergeProgressReports(batch, job)
	}

	i.ensureTimestamps(batch)
	i.cleanDirectoryStructure(batch)

	if ctx.Err() != nil {
		return fmt.Errorf("(io) %w: %w", ErrContextError, ctx.Err())
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
		return fmt.Errorf("(io) failed to check src in use: %w", err)
	} else if inUse {
		return fmt.Errorf("(io) %w", ErrSourceFileInUse)
	}

	if err := i.ensureDirectoryStructure(m, intermediateJob); err != nil {
		return fmt.Errorf("(io) failed to ensure dir structure: %w", err)
	}

	if !m.Metadata.IsDir && !m.IsHardlink && !m.IsSymlink && !m.Metadata.IsSymlink {
		if err := i.processFile(ctx, m); err != nil {
			return fmt.Errorf("(io) failed to process file: %w", err)
		}
		jobComplete = true
	}

	if m.Metadata.IsDir {
		if err := i.processDirectory(m); err != nil {
			return fmt.Errorf("(io) failed to process directory: %w", err)
		}
		jobComplete = true
	}

	if m.IsHardlink {
		if err := i.processHardlink(m); err != nil {
			return fmt.Errorf("(io) failed to process hardlink: %w", err)
		}
		jobComplete = true
	}

	if m.IsSymlink {
		if err := i.processSymlink(m, true); err != nil {
			return fmt.Errorf("(io) failed to process symlink: %w", err)
		}
		jobComplete = true
	}

	if m.Metadata.IsSymlink {
		if err := i.processSymlink(m, false); err != nil {
			return fmt.Errorf("(io) failed to process ext symlink: %w", err)
		}
		jobComplete = true
	}

	if !jobComplete {
		return fmt.Errorf("(io) %w", ErrNothingToProcess)
	}

	return nil
}
