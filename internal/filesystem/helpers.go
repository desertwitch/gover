package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/desertwitch/gover/internal/unraid"
	"golang.org/x/sys/unix"
)

type DiskStats struct {
	TotalSize uint64
	FreeSpace uint64
}

func (f *FileHandler) Exists(path string) (bool, error) {
	if _, err := f.OSOps.Stat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, fs.ErrNotExist
		}

		return false, err
	}

	return true, nil
}

func (f *FileHandler) ReadDir(name string) ([]os.DirEntry, error) {
	return f.OSOps.ReadDir(name)
}

func (f *FileHandler) ExistsOnStorage(m *Moveable) (string, error) {
	if m.Dest == nil {
		return "", ErrNilDestination
	}

	if _, ok := m.Dest.(*unraid.Disk); ok {
		for name, disk := range m.Share.IncludedDisks {
			if _, exists := m.Share.ExcludedDisks[name]; exists {
				continue
			}
			alreadyExists, existsPath, err := f.existsOnStorageCandidate(m, disk)
			if err != nil {
				return "", err
			}
			if alreadyExists {
				return existsPath, nil
			}
		}

		return "", nil
	}

	if pool, ok := m.Dest.(*unraid.Pool); ok {
		alreadyExists, existsPath, err := f.existsOnStorageCandidate(m, pool)
		if err != nil {
			return "", err
		}
		if alreadyExists {
			return existsPath, nil
		}

		return "", nil
	}

	return "", ErrImpossibleType
}

func (f *FileHandler) GetDiskUsage(path string) (DiskStats, error) {
	var stat unix.Statfs_t
	if err := f.UnixOps.Statfs(path, &stat); err != nil {
		return DiskStats{}, fmt.Errorf("failed to statfs: %w", err)
	}

	stats := DiskStats{
		TotalSize: stat.Blocks * handleSize(stat.Bsize),
		FreeSpace: stat.Bavail * handleSize(stat.Bsize),
	}

	return stats, nil
}

func (f *FileHandler) HasEnoughFreeSpace(s unraid.Storeable, minFree uint64, fileSize uint64) (bool, error) {
	path := s.GetFSPath()

	stats, err := f.GetDiskUsage(path)
	if err != nil {
		return false, fmt.Errorf("failed to get usage: %w", err)
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

func (f *FileHandler) IsEmptyFolder(path string) (bool, error) {
	entries, err := f.OSOps.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("failed to readdir: %w", err)
	}

	return len(entries) == 0, nil
}

func (f *FileHandler) existsOnStorageCandidate(m *Moveable, destCandidate unraid.Storeable) (bool, string, error) {
	relPath, err := filepath.Rel(m.Source.GetFSPath(), m.SourcePath)
	if err != nil {
		return false, "", fmt.Errorf("failed to rel path: %w", err)
	}

	dstPath := filepath.Join(destCandidate.GetFSPath(), relPath)

	if _, err := f.OSOps.Stat(dstPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, "", nil
		}

		return false, "", fmt.Errorf("failed to check existence: %w", err)
	}

	return true, dstPath, nil
}

func handleSize(size int64) uint64 {
	if size < 0 {
		return 0
	}

	return uint64(size)
}
