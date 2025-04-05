package filesystem

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/desertwitch/gover/internal/schema"
	"golang.org/x/sys/unix"
)

const (
	// DiskUsageCacherInterval is the updating interval of the disk usage cache.
	DiskUsageCacherInterval = 3 * time.Second
)

// unixStatfsProvider defines Statfs methods needed for disk usage checking.
type unixStatfsProvider interface {
	Statfs(path string, buf *unix.Statfs_t) error
}

// DiskUsageCacher caches disk usage information in a thread-safe manner.
type DiskUsageCacher struct {
	sync.RWMutex
	unixHandler unixStatfsProvider
	cache       map[string]*diskUsage
}

// DiskStats holds disk usage information. It is meant to be passed by value.
type DiskStats struct {
	TotalSize uint64
	FreeSpace uint64
}

// diskUsage holds [DiskStats] of a specific [schema.Storage].
type diskUsage struct {
	storage schema.Storage
	stats   DiskStats
}

// NewDiskUsageCacher returns a pointer to a new [DiskUsageCacher]. The update
// method is started, refreshing the cached data every
// [DiskUsageCacherInterval].
func NewDiskUsageCacher(ctx context.Context, unixHandler unixStatfsProvider) *DiskUsageCacher {
	cacher := &DiskUsageCacher{
		unixHandler: unixHandler,
		cache:       make(map[string]*diskUsage),
	}
	go cacher.periodicUpdate(ctx)

	return cacher
}

// getDiskUsageFromOS gets the actual [DiskStats] for a given path from the OS.
func (c *DiskUsageCacher) getDiskUsageFromOS(path string) (DiskStats, error) {
	var stat unix.Statfs_t
	if err := c.unixHandler.Statfs(path, &stat); err != nil {
		return DiskStats{}, fmt.Errorf("(fs-diskstats) failed to statfs: %w", err)
	}

	stats := DiskStats{
		TotalSize: stat.Blocks * handleSize(stat.Bsize),
		FreeSpace: stat.Bavail * handleSize(stat.Bsize),
	}

	return stats, nil
}

// periodicUpdate calls [DiskUsageCacher.Update] at a defined
// [DiskUsageCacherInterval].
func (c *DiskUsageCacher) periodicUpdate(ctx context.Context) {
	ticker := time.NewTicker(DiskUsageCacherInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = c.Update()
		}
	}
}

// Update calls [DiskUsageCacher.getDiskUsageFromOS] on all cached
// [schema.Storage]. This means that the cache is refreshed with new disk usage
// statistics from the OS.
func (c *DiskUsageCacher) Update() error {
	c.Lock()
	defer c.Unlock()

	for k, v := range c.cache {
		stats, err := c.getDiskUsageFromOS(v.storage.GetFSPath())
		if err != nil {
			delete(c.cache, k)

			return err
		}
		v.stats = stats
	}

	return nil
}

// GetDiskUsageFresh calls [DiskUsageCacher.getDiskUsageFromOS] for a specific
// [schema.Storage]. The result is stored in the [DiskUsageCacher]'s cache for
// later retrieval/periodic updating.
func (c *DiskUsageCacher) GetDiskUsageFresh(s schema.Storage) (DiskStats, error) {
	c.Lock()
	defer c.Unlock()

	stats, err := c.getDiskUsageFromOS(s.GetFSPath())
	if err != nil {
		return DiskStats{}, fmt.Errorf("(fs-diskstats-store) failed to get usage: %w", err)
	}

	c.cache[s.GetName()] = &diskUsage{
		storage: s,
		stats:   stats,
	}

	return stats, nil
}

// GetDiskUsage gets the cached [DiskStats] for a [schema.Storage]. If none are
// cached, fresh ones are retrieved and cached using
// [DiskUsageCacher.GetDiskUsageFresh].
func (c *DiskUsageCacher) GetDiskUsage(s schema.Storage) (DiskStats, error) {
	c.RLock()
	if cachedStats, exists := c.cache[s.GetName()]; exists {
		stats := cachedStats.stats
		c.RUnlock()

		return stats, nil
	}
	c.RUnlock()

	return c.GetDiskUsageFresh(s)
}

// HasEnoughFreeSpace is a helper method that allows checking if a certain
// [schema.Storage] can house a certain fileSize without exceeding a certain
// minFree threshold (uses caching).
func (c *DiskUsageCacher) HasEnoughFreeSpace(s schema.Storage, minFree uint64, fileSize uint64) (bool, error) {
	stats, err := c.GetDiskUsage(s)
	if err != nil {
		return false, fmt.Errorf("(fs-diskstats-efree) failed to get usage: %w", err)
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
