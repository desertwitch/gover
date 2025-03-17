package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/desertwitch/gover/internal/generic/schema"
)

func (f *Handler) IsInUse(path string) bool {
	return f.inUseHandler.IsInUse(path)
}

func (f *Handler) GetDiskUsage(s schema.Storage) (DiskStats, error) {
	return f.diskStatHandler.GetDiskUsage(s)
}

func (f *Handler) HasEnoughFreeSpace(s schema.Storage, minFree uint64, fileSize uint64) (bool, error) {
	return f.diskStatHandler.HasEnoughFreeSpace(s, minFree, fileSize)
}

func (f *Handler) ReadDir(name string) ([]os.DirEntry, error) {
	return f.osHandler.ReadDir(name)
}

func (f *Handler) IsEmptyFolder(path string) (bool, error) {
	entries, err := f.osHandler.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("(fs-isempty) failed to readdir: %w", err)
	}

	return len(entries) == 0, nil
}

func (f *Handler) Exists(path string) (bool, error) {
	if _, err := f.osHandler.Stat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, fs.ErrNotExist
		}

		return false, err
	}

	return true, nil
}

func (f *Handler) ExistsOnStorage(m *schema.Moveable) (string, error) {
	if m.Dest == nil {
		return "", ErrNilDestination
	}

	switch dest := m.Dest.(type) {
	case schema.Disk:
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

	case schema.Pool:
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

func (f *Handler) existsOnStorageCandidate(m *schema.Moveable, destCandidate schema.Storage) (bool, string, error) {
	relPath, err := filepath.Rel(m.Source.GetFSPath(), m.SourcePath)
	if err != nil {
		return false, "", fmt.Errorf("(fs-existson) failed to rel: %w", err)
	}

	dstPath := filepath.Join(destCandidate.GetFSPath(), relPath)

	if _, err := f.osHandler.Stat(dstPath); err != nil {
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

func concFilterSlice[T any](ctx context.Context, maxWorkers int, items []T, filterFunc func(T) bool) ([]T, error) {
	var wg sync.WaitGroup

	ch := make(chan T, len(items))

	wg.Add(1)
	go func() {
		defer wg.Done()

		semaphore := make(chan struct{}, maxWorkers)

		for _, item := range items {
			select {
			case <-ctx.Done():
				return
			case semaphore <- struct{}{}:
			}

			wg.Add(1)
			go func(item T) {
				defer wg.Done()
				defer func() { <-semaphore }()

				if filterFunc(item) {
					select {
					case <-ctx.Done():
						return
					case ch <- item:
					}
				}
			}(item)
		}
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	filtered := []T{}
	for item := range ch {
		filtered = append(filtered, item)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return filtered, nil
}
