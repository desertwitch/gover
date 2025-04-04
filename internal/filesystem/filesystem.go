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

type osProvider interface {
	Open(name string) (*os.File, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	ReadDir(name string) ([]os.DirEntry, error)
	Readlink(name string) (string, error)
	Remove(name string) error
	Rename(oldpath, newpath string) error
	Stat(name string) (os.FileInfo, error)
}

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

type inUseProvider interface {
	IsInUse(path string) bool
}

type fsWalkProvider interface {
	WalkDir(root string, fn fs.WalkDirFunc) error
}

type diskStatProvider interface {
	GetDiskUsage(storage schema.Storage) (DiskStats, error)
	HasEnoughFreeSpace(s schema.Storage, minFree uint64, fileSize uint64) (bool, error)
}

type Handler struct {
	osHandler       osProvider
	unixHandler     unixProvider
	inUseHandler    inUseProvider
	fileWalkHandler fsWalkProvider
	diskStatHandler diskStatProvider
}

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
