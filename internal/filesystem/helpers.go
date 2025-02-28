package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"

	"github.com/desertwitch/gover/internal/unraid"
	"golang.org/x/sys/unix"
)

type DiskStats struct {
	TotalSize int64
	FreeSpace int64
}

func (f *FilesystemImpl) ExistsOnStorage(m *Moveable) (storeable unraid.UnraidStoreable, existingAtPath string, err error) {
	if m.Dest == nil {
		return nil, "", fmt.Errorf("destination is nil")
	}

	if _, ok := m.Dest.(*unraid.UnraidDisk); ok {
		for name, disk := range m.Share.IncludedDisks {
			if _, exists := m.Share.ExcludedDisks[name]; exists {
				continue
			}
			alreadyExists, existsPath, err := existsOnStorageCandidate(m, disk, f.OSOps)
			if err != nil {
				return nil, "", err
			}
			if alreadyExists {
				return disk, existsPath, nil
			}
		}
		return nil, "", nil
	}

	if pool, ok := m.Dest.(*unraid.UnraidPool); ok {
		alreadyExists, existsPath, err := existsOnStorageCandidate(m, pool, f.OSOps)
		if err != nil {
			return nil, "", err
		}
		if alreadyExists {
			return pool, existsPath, nil
		}
		return nil, "", nil
	}

	return nil, "", fmt.Errorf("impossible storeable type")
}

func (f *FilesystemImpl) GetDiskUsage(path string) (DiskStats, error) {
	var stat unix.Statfs_t
	if err := f.UnixOps.Statfs(path, &stat); err != nil {
		return DiskStats{}, fmt.Errorf("failed to statfs: %w", err)
	}
	stats := DiskStats{
		TotalSize: int64(stat.Blocks) * int64(stat.Bsize),
		FreeSpace: int64(stat.Bavail) * int64(stat.Bsize),
	}
	return stats, nil
}

func (f *FilesystemImpl) HasEnoughFreeSpace(s unraid.UnraidStoreable, minFree int64, fileSize int64) (bool, error) {
	if fileSize < 0 {
		return false, fmt.Errorf("invalid file size < 0: %d", fileSize)
	}

	path := s.GetFSPath()

	stats, err := f.GetDiskUsage(path)
	if err != nil {
		return false, fmt.Errorf("failed to get usage: %w", err)
	}

	if stats.TotalSize <= 0 || stats.FreeSpace < 0 {
		return false, fmt.Errorf("invalid stats (TotalSize: %d, FreeSpace: %d)", stats.TotalSize, stats.FreeSpace)
	}

	requiredFree := minFree
	if minFree <= fileSize {
		requiredFree = fileSize
	}

	if stats.FreeSpace > requiredFree {
		return true, nil
	}

	return false, nil
}

func (f *FilesystemImpl) IsEmptyFolder(path string) (bool, error) {
	entries, err := f.OSOps.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("failed to readdir: %w", err)
	}
	return len(entries) == 0, nil
}

func (f *FilesystemImpl) IsFileInUse(path string) (bool, error) {
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

func existsOnStorageCandidate(m *Moveable, destCandidate unraid.UnraidStoreable, osOps osProvider) (exists bool, existingAtPath string, err error) {
	relPath, err := filepath.Rel(m.Source.GetFSPath(), m.SourcePath)
	if err != nil {
		return false, "", fmt.Errorf("failed to rel path: %w", err)
	}

	dstPath := filepath.Join(destCandidate.GetFSPath(), relPath)

	if _, err := osOps.Stat(dstPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check existence: %w", err)
	}

	return true, dstPath, nil
}
