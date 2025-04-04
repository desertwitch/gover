package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"

	"github.com/desertwitch/gover/internal/schema"
)

// IsInUse checks if a file is in use by another process on the operating system.
// For this it wraps the function of the given [inUseProvider] implementation.
func (f *Handler) IsInUse(path string) bool {
	return f.inUseHandler.IsInUse(path)
}

// GetDiskUsage gets [DiskStats] for a [schema.Storage], wrapping the
// previously given [diskStatProvider] implementation's respective function.
func (f *Handler) GetDiskUsage(s schema.Storage) (DiskStats, error) {
	data, err := f.diskStatHandler.GetDiskUsage(s)
	if err != nil {
		return data, fmt.Errorf("(fs-diskusage) %w", err)
	}

	return data, nil
}

// HasEnoughFreeSpace allows checking if a certain [schema.Storage] can house a
// certain fileSize without exceeding a certain minFree threshold. For this it
// wraps the function of the previously given [diskStatProvider] implementation.
func (f *Handler) HasEnoughFreeSpace(s schema.Storage, minFree uint64, fileSize uint64) (bool, error) {
	data, err := f.diskStatHandler.HasEnoughFreeSpace(s, minFree, fileSize)
	if err != nil {
		return data, fmt.Errorf("(fs-enoughspace) %w", err)
	}

	return data, nil
}

// IsEmptyFolder is a helper function checking if a path is an empty folder.
func (f *Handler) IsEmptyFolder(path string) (bool, error) {
	entries, err := f.osHandler.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("(fs-isempty) failed to readdir: %w", err)
	}

	return len(entries) == 0, nil
}

// Exists is a helper function checking if a path already exists.
func (f *Handler) Exists(path string) (bool, error) {
	if _, err := f.osHandler.Stat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, fs.ErrNotExist
		}

		return false, err
	}

	return true, nil
}

// ExistsOnStorage checks if a [schema.Moveable] path exists on the allocated storage.
//
// For the allocated destination as a [schema.Pool], it checks if the path exists only
// on that specific [schema.Pool].
//
// For the allocated destination as a [schema.Disk], it checks if the path exists on any
// of the [schema.Share]'s included disks (of an array), to avoid duplication when pooled.
func (f *Handler) ExistsOnStorage(m *schema.Moveable) (string, error) {
	if m.Dest == nil {
		return "", ErrNilDestination
	}

	switch dest := m.Dest.(type) {
	case schema.Disk:
		for _, disk := range m.Share.GetIncludedDisks() {
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

// existsOnStorageCandidate checks if a [schema.Moveable] path exists on a specific [schema.Storage].
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

// handleSize converts a int64 filesize to a uint64 filesize (with sizes < 0 becoming 0).
func handleSize(size int64) uint64 {
	if size < 0 {
		return 0
	}

	return uint64(size)
}

// concFilterSlice is a generic function concurrently filtering a slice using a given filtering function.
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
		return nil, fmt.Errorf("(fs-concfs) %w", ctx.Err())
	}

	return filtered, nil
}
