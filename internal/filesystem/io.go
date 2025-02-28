package filesystem

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"

	"github.com/desertwitch/gover/internal/unraid"
	"github.com/zeebo/blake3"
)

type relatedElement interface {
	GetMetadata() *Metadata
	GetSourcePath() string
	GetDestPath() string
}

type InternalProgressReport struct {
	AnyProcessed       []relatedElement
	DirsProcessed      []*RelatedDirectory
	MoveablesProcessed []*Moveable
	SymlinksProcessed  []*Moveable
	HardlinksProcessed []*Moveable
}

// TO-DO:
// Reallocation if not enough space (up to 3x?)
// Rollback, Locking?

func (f *FileHandler) ProcessMoveables(moveables []*Moveable, batch *InternalProgressReport) error {
	for _, m := range moveables {
		job := &InternalProgressReport{}

		if err := processMoveable(m, job, f, f.OSOps, f.UnixOps); err != nil {
			slog.Warn("Skipped job: failure during processing for job", "path", m.DestPath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}
		slog.Info("Processed:", "path", m.DestPath, "job", m.SourcePath, "share", m.Share.Name)

		for _, h := range m.Hardlinks {
			if err := processMoveable(h, job, f, f.OSOps, f.UnixOps); err != nil {
				slog.Warn("Skipped subjob: failure during processing for subjob", "path", h.DestPath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
				continue
			}
			slog.Info("Processed (hardlink):", "path", h.DestPath, "job", m.SourcePath, "share", m.Share.Name)
		}

		for _, s := range m.Symlinks {
			if err := processMoveable(s, job, f, f.OSOps, f.UnixOps); err != nil {
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

	if err := ensureTimestamps(batch, f.UnixOps); err != nil {
		return fmt.Errorf("failed finalizing timestamps: %w", err)
	}

	if err := removeEmptyDirs(batch, f, f.OSOps); err != nil {
		return fmt.Errorf("failed cleaning source directories: %w", err)
	}

	return nil
}

func processMoveable(m *Moveable, job *InternalProgressReport, fsOps fsProvider, osOps osProvider, unixOps unixProvider) error {
	used, err := fsOps.IsFileInUse(m.SourcePath)
	if err != nil {
		return fmt.Errorf("failed checking if source file is in use: %w", err)
	}
	if used {
		return fmt.Errorf("source file is currently in use")
	}

	if m.Hardlink {
		if err := ensureDirectoryStructure(m, job, osOps, unixOps); err != nil {
			return fmt.Errorf("failed to ensure dir tree for hardlink: %w", err)
		}

		if err := unixOps.Link(m.HardlinkTo.DestPath, m.DestPath); err != nil {
			return fmt.Errorf("failed to create hardlink: %w", err)
		}
		if err := osOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}

		if err := ensureLinkPermissions(m.DestPath, m.Metadata, unixOps); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.HardlinksProcessed = append(job.HardlinksProcessed, m)
		return nil
	}

	if m.Symlink {
		if err := ensureDirectoryStructure(m, job, osOps, unixOps); err != nil {
			return fmt.Errorf("failed to ensure dir tree for symlink: %w", err)
		}

		if err := unixOps.Symlink(m.SymlinkTo.DestPath, m.DestPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
		if err := osOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}

		if err := ensureLinkPermissions(m.DestPath, m.Metadata, unixOps); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.SymlinksProcessed = append(job.SymlinksProcessed, m)
		return nil
	}

	if m.Metadata.IsSymlink {
		if err := ensureDirectoryStructure(m, job, osOps, unixOps); err != nil {
			return fmt.Errorf("failed to ensure dir tree: %w", err)
		}

		if err := unixOps.Symlink(m.Metadata.SymlinkTo, m.DestPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
		if err := osOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}

		if err := ensureLinkPermissions(m.DestPath, m.Metadata, unixOps); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.MoveablesProcessed = append(job.MoveablesProcessed, m)
		return nil
	}

	if err := ensureDirectoryStructure(m, job, osOps, unixOps); err != nil {
		return fmt.Errorf("failed to ensure dir tree: %w", err)
	}

	if m.Metadata.IsDir {
		if err := unixOps.Mkdir(m.DestPath, m.Metadata.Perms); err != nil {
			return fmt.Errorf("failed to create empty dir: %w", err)
		}
		if err := osOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}
	} else {
		enoughSpace, err := fsOps.HasEnoughFreeSpace(m.Dest, m.Share.SpaceFloor, m.Metadata.Size)
		if err != nil {
			return fmt.Errorf("failed to check for enough space: %w", err)
		}
		if !enoughSpace {
			if _, ok := m.Dest.(*unraid.UnraidDisk); ok {
				// TO-DO: Reallocate with hardlinks
			} else {
				return fmt.Errorf("not enough free space on destination pool")
			}
		}

		if err := moveFile(m, osOps); err != nil {
			return fmt.Errorf("failed to move file: %w", err)
		}
		if err := osOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}
	}

	if err := ensurePermissions(m.DestPath, m.Metadata, unixOps); err != nil {
		return fmt.Errorf("failed to ensure permissions: %w", err)
	}

	job.AnyProcessed = append(job.AnyProcessed, m)
	job.MoveablesProcessed = append(job.MoveablesProcessed, m)
	return nil
}

func moveFile(m *Moveable, osOps osProvider) error {
	srcFile, err := osOps.Open(m.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	tmpPath := m.DestPath + ".gover"
	defer func() {
		if err != nil {
			osOps.Remove(tmpPath)
		}
	}()

	dstFile, err := osOps.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, os.FileMode(m.Metadata.Perms))
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

	srcChecksum := fmt.Sprintf("%x", srcHasher.Sum(nil))
	dstChecksum := fmt.Sprintf("%x", dstHasher.Sum(nil))

	if srcChecksum != dstChecksum {
		return fmt.Errorf("hash mismatch: %s (src) != %s (dst)", srcChecksum, dstChecksum)
	}

	if _, err := osOps.Stat(m.DestPath); err == nil {
		return fmt.Errorf("rename destination already exists")
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to check rename destination existence: %w", err)
	}

	if err := osOps.Rename(tmpPath, m.DestPath); err != nil {
		return fmt.Errorf("failed to rename temporary file to destination file: %w", err)
	}

	return nil
}
