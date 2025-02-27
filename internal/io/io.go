package io

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"sort"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
	"github.com/zeebo/blake3"
	"golang.org/x/sys/unix"
)

// TO-DO:
// Reallocation if not enough space (up to 3x?)
// Rollback, Locking?

func ProcessMoveables(moveables []*filesystem.Moveable, batch *InternalProgressReport) error {
	for _, m := range moveables {
		job := &InternalProgressReport{}

		if err := processMoveable(m, job); err != nil {
			slog.Warn("Skipped job: failure during processing for job", "path", m.DestPath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}
		slog.Info("Processed:", "path", m.DestPath, "job", m.SourcePath, "share", m.Share.Name)

		for _, h := range m.Hardlinks {
			if err := processMoveable(h, job); err != nil {
				slog.Warn("Skipped subjob: failure during processing for subjob", "path", h.DestPath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
				continue
			}
			slog.Info("Processed (hardlink):", "path", h.DestPath, "job", m.SourcePath, "share", m.Share.Name)
		}

		for _, s := range m.Symlinks {
			if err := processMoveable(s, job); err != nil {
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

	if err := ensureTimestamps(batch); err != nil {
		return fmt.Errorf("failed finalizing timestamps: %w", err)
	}

	if err := removeEmptyDirs(batch); err != nil {
		return fmt.Errorf("failed cleaning source directories: %w", err)
	}

	return nil
}

func processMoveable(m *filesystem.Moveable, job *InternalProgressReport) error {
	used, err := isFileInUse(m.SourcePath)
	if err != nil {
		return fmt.Errorf("failed checking if source file is in use: %w", err)
	}
	if used {
		return fmt.Errorf("source file is currently in use")
	}

	if m.Hardlink {
		if err := ensureDirectoryStructure(m, job); err != nil {
			return fmt.Errorf("failed to ensure dir tree for hardlink: %w", err)
		}

		if err := unix.Link(m.HardlinkTo.DestPath, m.DestPath); err != nil {
			return fmt.Errorf("failed to create hardlink: %w", err)
		}
		if err := os.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}

		if err := ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.HardlinksProcessed = append(job.HardlinksProcessed, m)
		return nil
	}

	if m.Symlink {
		if err := ensureDirectoryStructure(m, job); err != nil {
			return fmt.Errorf("failed to ensure dir tree for symlink: %w", err)
		}

		if err := unix.Symlink(m.SymlinkTo.DestPath, m.DestPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
		if err := os.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}

		if err := ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.SymlinksProcessed = append(job.SymlinksProcessed, m)
		return nil
	}

	if m.Metadata.IsSymlink {
		if err := ensureDirectoryStructure(m, job); err != nil {
			return fmt.Errorf("failed to ensure dir tree: %w", err)
		}

		if err := unix.Symlink(m.Metadata.SymlinkTo, m.DestPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
		if err := os.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}

		if err := ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}

		job.AnyProcessed = append(job.AnyProcessed, m)
		job.MoveablesProcessed = append(job.MoveablesProcessed, m)
		return nil
	}

	if err := ensureDirectoryStructure(m, job); err != nil {
		return fmt.Errorf("failed to ensure dir tree: %w", err)
	}

	if m.Metadata.IsDir {
		if err := unix.Mkdir(m.DestPath, m.Metadata.Perms); err != nil {
			return fmt.Errorf("failed to create empty dir: %w", err)
		}
		if err := os.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}
	} else {
		enoughSpace, err := filesystem.HasEnoughFreeSpace(m.Dest, m.Share.SpaceFloor, m.Metadata.Size)
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

		if err := moveFile(m); err != nil {
			return fmt.Errorf("failed to move file: %w", err)
		}
		if err := os.Remove(m.SourcePath); err != nil {
			return fmt.Errorf("failed to remove source after move: %w", err)
		}
	}

	if err := ensurePermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("failed to ensure permissions: %w", err)
	}

	job.AnyProcessed = append(job.AnyProcessed, m)
	job.MoveablesProcessed = append(job.MoveablesProcessed, m)
	return nil
}

func moveFile(m *filesystem.Moveable) error {
	srcFile, err := os.Open(m.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	tmpPath := m.DestPath + ".gover"
	defer func() {
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	dstFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, os.FileMode(m.Metadata.Perms))
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

	if _, err := os.Stat(m.DestPath); err == nil {
		return fmt.Errorf("rename destination already exists")
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to check rename destination existence: %w", err)
	}

	if err := os.Rename(tmpPath, m.DestPath); err != nil {
		return fmt.Errorf("failed to rename temporary file to destination file: %w", err)
	}

	return nil
}

func ensureTimestamps(batch *InternalProgressReport) error {
	for _, a := range batch.AnyProcessed {
		if err := ensureTimestamp(a.GetDestPath(), a.GetMetadata()); err != nil {
			slog.Warn("Warning (finalize): failure setting timestamp", "path", a.GetDestPath(), "err", err)
			continue
		}
	}
	return nil
}

func ensureTimestamp(path string, metadata *filesystem.Metadata) error {
	ts := []unix.Timespec{metadata.AccessedAt, metadata.ModifiedAt}
	if err := unix.UtimesNano(path, ts); err != nil {
		return fmt.Errorf("failed to set timestamp: %w", err)
	}
	return nil
}

func ensureDirectoryStructure(m *filesystem.Moveable, job *InternalProgressReport) error {
	dir := m.RootDir

	for dir != nil {
		// TO-DO: Handle generic errors here and otherwere for .Stat or .Lstat
		if _, err := os.Stat(dir.DestPath); errors.Is(err, fs.ErrNotExist) {
			if err := unix.Mkdir(dir.DestPath, dir.Metadata.Perms); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir.DestPath, err)
			}

			if err := ensurePermissions(dir.DestPath, dir.Metadata); err != nil {
				return fmt.Errorf("failed to ensure permissions: %w", err)
			}

			job.AnyProcessed = append(job.AnyProcessed, dir)
			job.DirsProcessed = append(job.DirsProcessed, dir)
		}
		dir = dir.Child
	}

	return nil
}

func ensurePermissions(path string, metadata *filesystem.Metadata) error {
	if err := unix.Chown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on %s: %w", path, err)
	}

	if err := unix.Chmod(path, metadata.Perms); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", path, err)
	}

	return nil
}

func ensureLinkPermissions(path string, metadata *filesystem.Metadata) error {
	if err := unix.Lchown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on link %s: %w", path, err)
	}

	return nil
}

func calculateDirectoryDepth(dir *filesystem.RelatedDirectory) int {
	depth := 0
	for dir != nil {
		dir = dir.Parent
		depth++
	}
	return depth
}

func removeEmptyDirs(batch *InternalProgressReport) error {
	sort.Slice(batch.DirsProcessed, func(i, j int) bool {
		return calculateDirectoryDepth(batch.DirsProcessed[i]) > calculateDirectoryDepth(batch.DirsProcessed[j])
	})

	removed := make(map[string]struct{})

	for _, dir := range batch.DirsProcessed {
		if _, alreadyRemoved := removed[dir.SourcePath]; alreadyRemoved {
			continue
		}
		isEmpty, err := filesystem.IsEmptyFolder(dir.SourcePath)
		if err != nil {
			slog.Warn("Warning (cleanup): failure establishing source directory emptiness (skipped)", "path", dir.SourcePath, "err", err)
			continue
		}
		if isEmpty {
			if err := os.Remove(dir.SourcePath); err != nil {
				slog.Warn("Warning (cleanup): failure removing empty source directory (skipped)", "path", dir.SourcePath, "err", err)
				continue
			}
			removed[dir.SourcePath] = struct{}{}
		}
	}

	return nil
}

func isFileInUse(path string) (bool, error) {
	cmd := exec.Command("lsof", path)

	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return false, nil
	}

	return false, err
}
