package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/desertwitch/gover/internal/generic/storage"
	"golang.org/x/sys/unix"
)

type DiskStats struct {
	TotalSize uint64
	FreeSpace uint64
}

func (f *Handler) Exists(path string) (bool, error) {
	if _, err := f.OSHandler.Stat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, fs.ErrNotExist
		}

		return false, err
	}

	return true, nil
}

func (f *Handler) ReadDir(name string) ([]os.DirEntry, error) {
	return f.OSHandler.ReadDir(name)
}

func (f *Handler) ExistsOnStorage(m *Moveable) (string, error) {
	if m.Dest == nil {
		return "", ErrNilDestination
	}

	switch dest := m.Dest.(type) {
	case storage.Disk:
		for name, disk := range m.Share.GetIncludedDisks() {
			if _, exists := m.Share.GetExcludedDisks()[name]; exists {
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

	case storage.Pool:
		alreadyExists, existsPath, err := f.existsOnStorageCandidate(m, dest)
		if err != nil {
			return "", err
		}
		if alreadyExists {
			return existsPath, nil
		}

		return "", nil

	default:
		return "", ErrImpossibleType
	}
}

func (f *Handler) GetDiskUsage(path string) (DiskStats, error) {
	var stat unix.Statfs_t
	if err := f.UnixHandler.Statfs(path, &stat); err != nil {
		return DiskStats{}, fmt.Errorf("(fs-diskuse) failed to statfs: %w", err)
	}

	stats := DiskStats{
		TotalSize: stat.Blocks * handleSize(stat.Bsize),
		FreeSpace: stat.Bavail * handleSize(stat.Bsize),
	}

	return stats, nil
}

func (f *Handler) HasEnoughFreeSpace(s storage.Storage, minFree uint64, fileSize uint64) (bool, error) {
	path := s.GetFSPath()

	stats, err := f.GetDiskUsage(path)
	if err != nil {
		return false, fmt.Errorf("(fs-enoughfree) failed to get usage: %w", err)
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

func (f *Handler) IsEmptyFolder(path string) (bool, error) {
	entries, err := f.OSHandler.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("(fs-isempty) failed to readdir: %w", err)
	}

	return len(entries) == 0, nil
}

func (f *Handler) IsFileInUse(targetFile string) (bool, error) {
	procEntries, err := f.OSHandler.ReadDir("/proc")
	if err != nil {
		return false, fmt.Errorf("failed to read /proc: %w", err)
	}

	for _, procEntry := range procEntries {
		pid, err := strconv.Atoi(procEntry.Name())
		if err != nil {
			continue
		}

		fdPath := fmt.Sprintf("/proc/%d/fd", pid)
		fdEntries, err := f.OSHandler.ReadDir(fdPath)
		if err != nil {
			continue
		}

		for _, fdEntry := range fdEntries {
			fdLink := fmt.Sprintf("/proc/%d/fd/%s", pid, fdEntry.Name())
			linkTarget, err := f.OSHandler.Readlink(fdLink)
			if err != nil {
				continue
			}

			if linkTarget == targetFile {
				return true, nil
			}
		}
	}

	return false, nil
}

func (f *Handler) existsOnStorageCandidate(m *Moveable, destCandidate storage.Storage) (bool, string, error) {
	relPath, err := filepath.Rel(m.Source.GetFSPath(), m.SourcePath)
	if err != nil {
		return false, "", fmt.Errorf("(fs-existson) failed to rel: %w", err)
	}

	dstPath := filepath.Join(destCandidate.GetFSPath(), relPath)

	if _, err := f.OSHandler.Stat(dstPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, "", nil
		}

		return false, "", fmt.Errorf("(fs-existson) failed to stat: %w", err)
	}

	return true, dstPath, nil
}

func handleSize(size int64) uint64 {
	if size < 0 {
		return 0
	}

	return uint64(size)
}
