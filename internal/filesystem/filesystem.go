// Package filesystem implements routines for translating filesystem elements
// into [schema.Moveable] (by walking the filesystem), as well as helper
// routines relating to information collection on associated filesystems.
package filesystem

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/desertwitch/gover/internal/schema"
	"golang.org/x/sys/unix"
)

// osProvider defines operating system methods needed to work on the filesystem
// within the functions and methods of this package.
type osProvider interface {
	Open(name string) (*os.File, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	ReadDir(name string) ([]os.DirEntry, error)
	Readlink(name string) (string, error)
	Remove(name string) error
	Rename(oldpath, newpath string) error
	Stat(name string) (os.FileInfo, error)
}

// unixProvider defines Unix operating system methods needed to work on the
// filesystem within the functions and methods of this package.
type unixProvider interface {
	Chmod(path string, mode uint32) error
	Chown(path string, uid, gid int) error
	Lchown(path string, uid, gid int) error
	Link(oldpath, newpath string) error
	Lstat(path string, stat *unix.Stat_t) error
	Mkdir(path string, mode uint32) error
	Statfs(path string, buf *unix.Statfs_t) error
	Symlink(oldpath, newpath string) error
	UtimesNano(path string, times []unix.Timespec) error
}

// inUseProvider defines methods needed to check if a source path is presently
// in use by another process of the operating system.
type inUseProvider interface {
	IsInUse(path string) bool
}

// fsWalkProvider defines methods needed to traverse the filesystem.
type fsWalkProvider interface {
	WalkDir(root string, fn fs.WalkDirFunc) error
}

// diskStatProvider defines methods needed for disk usage statistics.
type diskStatProvider interface {
	GetDiskUsage(storage schema.Storage) (DiskStats, error)
	HasEnoughFreeSpace(s schema.Storage, minFree uint64, fileSize uint64) (bool, error)
}

// Handler is the principal implementation for the filesystem services.
type Handler struct {
	osHandler       osProvider
	unixHandler     unixProvider
	inUseHandler    inUseProvider
	fileWalkHandler fsWalkProvider
	diskStatHandler diskStatProvider
}

// NewHandler returns a pointer to a new filesystem [Handler].
func NewHandler(ctx context.Context, osHandler osProvider, unixHandler unixProvider) (*Handler, error) {
	inUseHandler, err := NewInUseChecker(ctx, osHandler)
	if err != nil {
		return nil, fmt.Errorf("(fs) failed to spawn file-in-use checker: %w", err)
	}

	fileWalkHandler := newFileWalker()
	diskStatHandler := NewDiskUsageCacher(ctx, unixHandler)

	return &Handler{
		osHandler:       osHandler,
		unixHandler:     unixHandler,
		inUseHandler:    inUseHandler,
		fileWalkHandler: fileWalkHandler,
		diskStatHandler: diskStatHandler,
	}, nil
}

// GetMoveables returns all [schema.Moveable] candidates for a [schema.Share] on
// a [schema.Storage].
//
// It is the principal method used for retrieving all [schema.Moveable] and
// their subelements for a [schema.Share] on a [schema.Storage], including
// referencing any hard- and symlinks, establishing metadata, as well as
// directory structure with parent/child relations for later
// allocation/recreation.
//
// For convenience, a destination [schema.Storage] can be set here, if it is
// already known at the time. An example case would be directly allocating to
// one [schema.Pool] instead of multiple [schema.Disk].
func (f *Handler) GetMoveables(ctx context.Context, share schema.Share, src schema.Storage, dst schema.Storage) ([]*schema.Moveable, error) {
	moveables := []*schema.Moveable{}

	shareDir := filepath.Join(src.GetFSPath(), share.GetName())

	err := f.fileWalkHandler.WalkDir(shareDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if path != shareDir {
				slog.Warn("Failure for path during walking of directory tree (was skipped)",
					"path", path,
					"err", err,
					"share", share.GetName(),
				)
			}

			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		isEmptyDir := false
		if d.IsDir() {
			isEmptyDir, err = f.IsEmptyFolder(path)
			if err != nil {
				slog.Warn("Failure checking for emptiness during walking of directory tree (was skipped)",
					"path", path,
					"err", err,
					"share", share.GetName(),
				)

				return nil
			}
		}

		if !d.IsDir() || (d.IsDir() && isEmptyDir) {
			moveable := &schema.Moveable{
				Share:      share,
				Source:     src,
				SourcePath: path,
				Dest:       dst,
			}

			moveables = append(moveables, moveable)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("(fs) failed walking: %w", err)
	}

	filtered, err := concFilterSlice(ctx, runtime.NumCPU(), moveables, func(m *schema.Moveable) bool {
		if err := f.establishMetadata(m); err != nil {
			return false
		}
		if err := f.establishRelatedDirs(m, shareDir); err != nil {
			return false
		}

		return true
	})
	if err != nil {
		return nil, fmt.Errorf("(fs) failed relating metadata: %w", err)
	}

	establishSymlinks(filtered, dst)
	establishHardlinks(filtered, dst)

	filtered = removeInternalLinks(filtered)
	filtered = f.removeInUseFiles(filtered)

	return filtered, nil
}

// removeInUseFiles removes from a slice of [schema.Moveable] these which are
// currently in use by another process of the operating system. For this, the
// previously given to the [Handler] implementation of [inUseProvider] is used.
func (f *Handler) removeInUseFiles(moveables []*schema.Moveable) []*schema.Moveable {
	filtered := []*schema.Moveable{}

	for _, m := range moveables {
		if !m.Metadata.IsDir {
			if inUse := f.inUseHandler.IsInUse(m.SourcePath); inUse {
				slog.Warn("Skipped job: source file is in use",
					"job", m.SourcePath,
					"share", m.Share.GetName(),
				)

				continue
			}
		}
		filtered = append(filtered, m)
	}

	return filtered
}
