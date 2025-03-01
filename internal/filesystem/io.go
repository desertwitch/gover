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

		if err := f.processMoveable(m, job); err != nil {
			slog.Warn("Skipped job: failure during processing for job", "path", m.DestPath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}
		slog.Info("Processed:", "path", m.DestPath, "job", m.SourcePath, "share", m.Share.Name)

		for _, h := range m.Hardlinks {
			if err := f.processMoveable(h, job); err != nil {
				slog.Warn("Skipped subjob: failure during processing for subjob", "path", h.DestPath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
				continue
			}
			slog.Info("Processed (hardlink):", "path", h.DestPath, "job", m.SourcePath, "share", m.Share.Name)
		}

		for _, s := range m.Symlinks {
			if err := f.processMoveable(s, job); err != nil {
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

	if err := f.ensureTimestamps(batch); err != nil {
		return fmt.Errorf("failed finalizing timestamps: %w", err)
	}

	if err := f.removeEmptyDirs(batch); err != nil {
		return fmt.Errorf("failed cleaning source directories: %w", err)
	}

	return nil
}

func (f *FileHandler) processMoveable(m *Moveable, job *InternalProgressReport) error {
	used, err := f.IsFileInUse(m.SourcePath)
	if err != nil {
		return fmt.Errorf("failed checking if source file is in use: %w", err)
	}
	if used {
		return fmt.Errorf("source file is currently in use")
	}

	if m.Hardlink {
		if err := f.ensureDirectoryStructure(m, job); err != nil {
			return fmt.Errorf("failed to ensure dir tree for hardlink: %w", err)
		}

		if err := f.UnixOps.Link(m.HardlinkTo.DestPath, m.DestPath); err != nil {
			return fmt.Errorf("failed to create hardlink: %w", err)
		}
		if err := f.OSOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}

		if err := f.ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.HardlinksProcessed = append(job.HardlinksProcessed, m)
		return nil
	}

	if m.Symlink {
		if err := f.ensureDirectoryStructure(m, job); err != nil {
			return fmt.Errorf("failed to ensure dir tree for symlink: %w", err)
		}

		if err := f.UnixOps.Symlink(m.SymlinkTo.DestPath, m.DestPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
		if err := f.OSOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}

		if err := f.ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.SymlinksProcessed = append(job.SymlinksProcessed, m)
		return nil
	}

	if m.Metadata.IsSymlink {
		if err := f.ensureDirectoryStructure(m, job); err != nil {
			return fmt.Errorf("failed to ensure dir tree: %w", err)
		}

		if err := f.UnixOps.Symlink(m.Metadata.SymlinkTo, m.DestPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
		if err := f.OSOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}

		if err := f.ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.MoveablesProcessed = append(job.MoveablesProcessed, m)
		return nil
	}

	if err := f.ensureDirectoryStructure(m, job); err != nil {
		return fmt.Errorf("failed to ensure dir tree: %w", err)
	}

	if m.Metadata.IsDir {
		if err := f.UnixOps.Mkdir(m.DestPath, m.Metadata.Perms); err != nil {
			return fmt.Errorf("failed to create empty dir: %w", err)
		}
		if err := f.OSOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}
	} else {
		enoughSpace, err := f.HasEnoughFreeSpace(m.Dest, m.Share.SpaceFloor, m.Metadata.Size)
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

		if err := f.moveFile(m); err != nil {
			return fmt.Errorf("failed to move file: %w", err)
		}
		if err := f.OSOps.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}
	}

	if err := f.ensurePermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("failed to ensure permissions: %w", err)
	}

	job.AnyProcessed = append(job.AnyProcessed, m)
	job.MoveablesProcessed = append(job.MoveablesProcessed, m)
	return nil
}

func (f *FileHandler) moveFile(m *Moveable) error {
	srcFile, err := f.OSOps.Open(m.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	tmpPath := m.DestPath + ".gover"
	defer func() {
		if err != nil {
			f.OSOps.Remove(tmpPath)
		}
	}()

	dstFile, err := f.OSOps.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, os.FileMode(m.Metadata.Perms))
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

	if _, err := f.OSOps.Stat(m.DestPath); err == nil {
		return fmt.Errorf("rename destination already exists")
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to check rename destination existence: %w", err)
	}

	if err := f.OSOps.Rename(tmpPath, m.DestPath); err != nil {
		return fmt.Errorf("failed to rename temporary file to destination file: %w", err)
	}

	return nil
}
