package io

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/storage"
	"golang.org/x/sys/unix"
)

type fsProvider interface {
	HasEnoughFreeSpace(s storage.Storage, minFree uint64, fileSize uint64) (bool, error)
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

type Handler struct {
	sync.Mutex
	FSHandler   fsProvider
	OSHandler   osProvider
	UnixHandler unixProvider
}

func NewHandler(fsHandler fsProvider, osHandler osProvider, unixHandler unixProvider) *Handler {
	return &Handler{
		FSHandler:   fsHandler,
		OSHandler:   osHandler,
		UnixHandler: unixHandler,
	}
}

func (i *Handler) ProcessQueue(ctx context.Context, q *queue.DestinationQueue) {
	batch := &creationReport{}

	for {
		if ctx.Err() != nil {
			break
		}

		m, ok := q.Dequeue()
		if !ok {
			break
		}

		job := &creationReport{}

		if err := i.processQueueElement(ctx, m, q, job); err != nil {
			continue
		}

		for _, h := range m.Hardlinks {
			if err := i.processQueueSubElement(ctx, h, m, q, job); err != nil {
				continue
			}
		}

		for _, s := range m.Symlinks {
			if err := i.processQueueSubElement(ctx, s, m, q, job); err != nil {
				continue
			}
		}

		mergeCreationReports(batch, job)
	}

	i.ensureTimestamps(batch)
	i.cleanDirectoryStructure(batch)
}

func (i *Handler) processQueueElement(ctx context.Context, elem *filesystem.Moveable, q *queue.DestinationQueue, job *creationReport) error {
	q.SetProcessing(elem)

	if err := i.processMoveable(ctx, elem, job); err != nil {
		slog.Warn("Skipped job: failure during processing",
			"path", elem.DestPath,
			"err", err,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)
		q.SetSkipped(elem)

		return err
	}

	q.SetSuccess(elem)

	slog.Info("Processed:",
		"path", elem.DestPath,
		"job", elem.SourcePath,
		"share", elem.Share.GetName(),
	)

	return nil
}

func (i *Handler) processQueueSubElement(ctx context.Context, subelem *filesystem.Moveable, elem *filesystem.Moveable, q *queue.DestinationQueue, job *creationReport) error {
	q.SetProcessing(subelem)

	if err := i.processMoveable(ctx, subelem, job); err != nil {
		slog.Warn("Skipped subjob: failure during processing",
			"path", subelem.DestPath,
			"err", err,
			"subjob", subelem.SourcePath,
			"job", elem.SourcePath,
			"share", elem.Share.GetName(),
		)
		q.SetSkipped(subelem)

		return err
	}

	q.SetSuccess(subelem)

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

func (i *Handler) processMoveable(ctx context.Context, m *filesystem.Moveable, job *creationReport) error {
	var jobComplete bool

	intermediateJob := &creationReport{}

	defer func() {
		if jobComplete {
			addToCreationReport(intermediateJob, m)
			mergeCreationReports(job, intermediateJob)
		} else {
			i.cleanDirectoriesAfterFailure(intermediateJob)
		}
	}()

	// if inUse, err := i.FSHandler.IsFileInUse(m.SourcePath); err != nil {
	// 	return fmt.Errorf("(io) failed to check src in use: %w", err)
	// } else if inUse {
	// 	return fmt.Errorf("(io) %w", ErrSourceFileInUse)
	// }

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
