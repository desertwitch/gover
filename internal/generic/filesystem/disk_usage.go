package filesystem

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/desertwitch/gover/internal/generic/schema"
	"golang.org/x/sys/unix"
)

const (
	DiskUsageCacherInterval = 3 * time.Second
)

type unixStatfsProvider interface {
	Statfs(path string, buf *unix.Statfs_t) error
}

type DiskUsageCacher struct {
	sync.RWMutex
	unixHandler unixStatfsProvider
	cache       map[string]*diskUsage
}

type DiskStats struct {
	TotalSize uint64
	FreeSpace uint64
}

type diskUsage struct {
	storage schema.Storage
	stats   DiskStats
}

func NewDiskUsageCacher(ctx context.Context, unixHandler unixStatfsProvider) *DiskUsageCacher {
	cacher := &DiskUsageCacher{
		unixHandler: unixHandler,
		cache:       make(map[string]*diskUsage),
	}
	go cacher.periodicUpdate(ctx)

	return cacher
}

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
