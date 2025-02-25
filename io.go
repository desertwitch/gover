package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

// TO-DO:
// Reallocation if not enough space (up to 3x?)
// Timestamp tracking at the end ranging over processed
// Rollback, JobProgress, Locking?

func processMoveables(moveables []*Moveable, p *BatchProgress) error {
	for _, m := range moveables {
		if err := processMoveable(m, p); err != nil {
			slog.Warn("Skipped job: failure during processing for job", "path", m.DestPath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}
		for _, h := range m.Hardlinks {
			if err := processMoveable(h, p); err != nil {
				slog.Warn("Skipped job: failure during processing for subjob", "path", h.DestPath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
				// TO-DO: Rollback?
				continue
			}
		}
		for _, s := range m.Symlinks {
			if err := processMoveable(s, p); err != nil {
				slog.Warn("Skipped job: failure during processing for subjob", "path", s.DestPath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
				// TO-DO: Rollback?
				continue
			}
		}
	}
	return nil
}

func processMoveable(m *Moveable, p *BatchProgress) error {
	used, err := isFileInUse(m.SourcePath)
	if err != nil {
		return fmt.Errorf("failed checking if source file is in use: %w", err)
	}
	if used {
		return fmt.Errorf("source file is currently in use")
	}

	if m.Hardlink {
		if err := ensureDirectoryStructure(m, p); err != nil {
			return fmt.Errorf("failed to ensure dir tree for hardlink: %w", err)
		}

		if err := unix.Link(m.HardlinkTo.DestPath, m.DestPath); err != nil {
			return fmt.Errorf("failed to create hardlink: %w", err)
		}

		if err := ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}
		p.AnyProcessed = append(p.AnyProcessed, m)
		p.HardlinksProcessed = append(p.HardlinksProcessed, m)

		return nil
	}

	if m.Symlink {
		if err := ensureDirectoryStructure(m, p); err != nil {
			return fmt.Errorf("failed to ensure dir tree for symlink: %w", err)
		}

		if err := unix.Symlink(m.SymlinkTo.DestPath, m.DestPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}

		if err := ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}
		p.AnyProcessed = append(p.AnyProcessed, m)
		p.SymlinksProcessed = append(p.SymlinksProcessed, m)

		return nil
	}

	if m.Metadata.IsSymlink {
		if err := ensureDirectoryStructure(m, p); err != nil {
			return fmt.Errorf("failed to ensure dir tree: %w", err)
		}

		if err := unix.Symlink(m.Metadata.SymlinkTo, m.DestPath); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}

		if err := ensureLinkPermissions(m.DestPath, m.Metadata); err != nil {
			return fmt.Errorf("failed to ensure link permissions: %w", err)
		}
		p.AnyProcessed = append(p.AnyProcessed, m)
		p.MoveablesProcessed = append(p.MoveablesProcessed, m)

		return nil
	}

	if err := ensureDirectoryStructure(m, p); err != nil {
		return fmt.Errorf("failed to ensure dir tree: %w", err)
	}

	if m.Metadata.IsDir {
		if err := unix.Mkdir(m.DestPath, m.Metadata.Perms); err != nil {
			return fmt.Errorf("failed to create empty dir: %w", err)
		}
	} else {
		enoughSpace, err := hasEnoughFreeSpace(m.Dest, m.Share.SpaceFloor, m.Metadata.Size)
		if err != nil {
			return fmt.Errorf("failed to check for enough space: %w", err)
		}
		if !enoughSpace {
			if _, ok := m.Dest.(*UnraidDisk); ok {
				// TO-DO: Reallocate with hardlinks
			} else {
				return fmt.Errorf("not enough free space on destination pool")
			}
		}

		if err := moveFile(m); err != nil {
			return fmt.Errorf("failed to move file: %w", err)
		}
	}

	if err := ensurePermissions(m.DestPath, m.Metadata); err != nil {
		return fmt.Errorf("failed to ensure permissions: %w", err)
	}
	p.AnyProcessed = append(p.AnyProcessed, m)
	p.MoveablesProcessed = append(p.MoveablesProcessed, m)

	return nil
}

func moveFile(m *Moveable) error {
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

	srcHasher := sha256.New()
	dstHasher := sha256.New()

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

func ensureDirectoryStructure(m *Moveable, p *BatchProgress) error {
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

			p.AnyProcessed = append(p.AnyProcessed, dir)
			p.DirsProcessed = append(p.DirsProcessed, dir)
		}
		dir = dir.Child
	}
	return nil
}

func ensurePermissions(path string, metadata *Metadata) error {
	if err := unix.Chown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on %s: %w", path, err)
	}

	if err := unix.Chmod(path, metadata.Perms); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", path, err)
	}

	return nil
}

func ensureLinkPermissions(path string, metadata *Metadata) error {
	if err := unix.Lchown(path, int(metadata.UID), int(metadata.GID)); err != nil {
		return fmt.Errorf("failed to set ownership on link %s: %w", path, err)
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
