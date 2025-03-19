package io

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/schema"
	"golang.org/x/sys/unix"
)

type fsProvider interface {
	HasEnoughFreeSpace(s schema.Storage, minFree uint64, fileSize uint64) (bool, error)
	IsEmptyFolder(path string) (bool, error)
	IsInUse(path string) bool
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

type ioTargetQueue interface {
	DequeueAndProcess(ctx context.Context, processFunc func(*schema.Moveable) int) error
}

type fsElement interface {
	GetDestPath() string
	GetMetadata() *schema.Metadata
	GetSourcePath() string
}

type Handler struct {
	sync.Mutex
	fsHandler   fsProvider
	osHandler   osProvider
	unixHandler unixProvider
}

func NewHandler(fsHandler fsProvider, osHandler osProvider, unixHandler unixProvider) *Handler {
	return &Handler{
		fsHandler:   fsHandler,
		osHandler:   osHandler,
		unixHandler: unixHandler,
	}
}

func (i *Handler) ProcessQueue(ctx context.Context, q ioTargetQueue) {
	batch := &ioReport{}

	q.DequeueAndProcess(ctx, func(m *schema.Moveable) int {
		job := &ioReport{}

		if err := i.processElement(ctx, m, job); err != nil {
			return queue.DecisionSkipped
		}

		for _, h := range m.Hardlinks {
			if err := i.processSubElement(ctx, h, m, job); err != nil {
				continue
			}
		}

		for _, s := range m.Symlinks {
			if err := i.processSubElement(ctx, s, m, job); err != nil {
				continue
			}
		}

		mergeIOReports(batch, job)

		return queue.DecisionSuccess
	})

	i.ensureTimestamps(batch)
	i.cleanDirectoryStructure(batch)
}

func (i *Handler) processElement(ctx context.Context, elem *schema.Moveable, job *ioReport) error {
	if err := i.processMoveable(ctx, elem, job); err != nil {
		slog.Warn("Skipped job: failure during processing",
			"path", elem.DestPath,
			"err", err,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return err
	}

	slog.Info("Processed:",
		"path", elem.DestPath,
		"job", elem.SourcePath,
		"share", elem.Share.GetName(),
	)

	return nil
}

func (i *Handler) processSubElement(ctx context.Context, subelem *schema.Moveable, elem *schema.Moveable, job *ioReport) error {
	if err := i.processMoveable(ctx, subelem, job); err != nil {
		slog.Warn("Skipped subjob: failure during processing",
			"path", subelem.DestPath,
			"err", err,
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)

		return err
	}

	linkType := "hardlink"
	if subelem.IsSymlink || subelem.Metadata.IsSymlink {
		linkType = "symlink"
	}

	slog.Info(fmt.Sprintf("Processed (%s):", linkType),
		"path", subelem.DestPath,
		"subjob", subelem.SourcePath,
		"job", elem.SourcePath,
		"share", elem.Share.GetName(),
	)

	return nil
}

func (i *Handler) processMoveable(ctx context.Context, m *schema.Moveable, job *ioReport) error {
	var jobComplete bool

	intermediateJob := &ioReport{}

	defer func() {
		if jobComplete {
			addToIOReport(intermediateJob, m)
			mergeIOReports(job, intermediateJob)
		} else {
			i.cleanDirectoriesAfterFailure(intermediateJob)
		}
	}()

	if inUse := i.fsHandler.IsInUse(m.SourcePath); inUse {
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
