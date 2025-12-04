// Package io implements routines for moving [schema.Moveable] between
// [schema.Storage]. It handles all filesystem manipulations and is designed to
// interact closely with the package [queue] and its IO (target) queues.
package io

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/desertwitch/gover/internal/queue"
	"github.com/desertwitch/gover/internal/schema"
	"golang.org/x/sys/unix"
)

// fsProvider defines the filesystem methods needed for IO operations.
type fsProvider interface {
	HasEnoughFreeSpace(s schema.Storage, minFree uint64, fileSize uint64) (bool, error)
	IsEmptyFolder(path string) (bool, error)
	IsInUse(path string) bool
}

// osProvider defines the operating system methods needed for IO operations.
type osProvider interface {
	Open(name string) (*os.File, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Remove(name string) error
	Rename(oldpath, newpath string) error
	Stat(name string) (os.FileInfo, error)
}

// unixProvider defines the Unix operating system methods needed for IO
// operations.
type unixProvider interface {
	Chmod(path string, mode uint32) error
	Chown(path string, uid, gid int) error
	Lchown(path string, uid, gid int) error
	Link(oldpath, newpath string) error
	Mkdir(path string, mode uint32) error
	Symlink(oldpath, newpath string) error
	UtimesNano(path string, times []unix.Timespec) error
}

// ioTargetQueue defines the methods an IO queue needs to have for IO
// operations.
type ioTargetQueue interface {
	AddBytesTransfered(bytes uint64)
	DequeueAndProcess(ctx context.Context, processFunc func(*schema.Moveable) int) error
	PreProcess(p schema.Pipeline[*schema.Moveable]) bool
	PostProcess(p schema.Pipeline[*schema.Moveable]) bool
}

// fsElement defines the methods any filesystem element needs to have for IO
// operations.
type fsElement interface {
	GetDestPath() string
	GetMetadata() *schema.Metadata
	GetSourcePath() string
}

// Handler is the principal implementation for the IO services. It is safe for
// concurrent use on grouped by target [schema.Storage] queues of
// [schema.Moveable].
type Handler struct {
	sync.Mutex

	fsHandler   fsProvider
	osHandler   osProvider
	unixHandler unixProvider
}

// NewHandler returns a pointer to a new IO [Handler].
func NewHandler(fsHandler fsProvider, osHandler osProvider, unixHandler unixProvider) *Handler {
	return &Handler{
		fsHandler:   fsHandler,
		osHandler:   osHandler,
		unixHandler: unixHandler,
	}
}

// ProcessTargetQueue sequentially processes an [ioTargetQueue], containing
// [schema.Moveable] grouped by one respective destination [schema.Storage].
//
// This method does not concurrently operate within a single [ioTargetQueue].
// Hence this function is usually called on multiple [ioTargetQueue]
// concurrently, but processing each respective [schema.Storage] in sequence.
func (i *Handler) ProcessTargetQueue(
	ctx context.Context,
	pipelines map[string]schema.Pipeline[*schema.Moveable],
	target schema.Storage,
	targetQueue ioTargetQueue,
) bool {
	batch := &ioReport{}

	defer func() {
		i.ensureTimestamps(batch)
		i.cleanDirectoryStructure(batch)
	}()

	if pipeline, exists := pipelines[target.GetName()]; exists {
		if success := targetQueue.PreProcess(pipeline); !success {
			slog.Warn("Skipped target storage: pre-processing pipeline failure",
				"target", target.GetName(),
			)

			return false
		}
	}

	if err := targetQueue.DequeueAndProcess(ctx, func(m *schema.Moveable) int {
		job := &ioReport{}

		if pipeline, exists := pipelines[target.GetName()]; exists {
			if success := pipeline.Process(m); !success {
				return queue.DecisionSkipped
			}
		}

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
		targetQueue.AddBytesTransfered(m.Metadata.Size)

		return queue.DecisionSuccess
	}); err != nil {
		return false
	}

	if pipeline, exists := pipelines[target.GetName()]; exists {
		if success := targetQueue.PostProcess(pipeline); !success {
			slog.Warn("Partial failure for target storage: post-processing pipeline failure",
				"target", target.GetName(),
			)

			return false
		}
	}

	return true
}

// processElement processes a dequeued "parent" [schema.Moveable] element.
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

// processSubElement processes a dequeued "child" [schema.Moveable]
// (hard-/symlink) subelement.
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

// processMoveable is the principal method for IO-processing a [schema.Moveable]
// of any supported type.
func (i *Handler) processMoveable(ctx context.Context, m *schema.Moveable, job *ioReport) error {
	var jobComplete bool

	intermediateJob := &ioReport{}

	defer func() {
		if jobComplete {
			addToIOReport(intermediateJob, m)
			mergeIOReports(job, intermediateJob)
		} else {
			i.cleanFileAfterFailure(m)
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
