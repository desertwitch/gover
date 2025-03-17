package allocation

import (
	"fmt"
	"sync"
	"time"

	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/schema"
)

type fsUsageProvider interface {
	GetDiskUsage(path string) (filesystem.DiskStats, error)
}

type DiskUsageCacher struct {
	sync.Mutex
	fsHandler fsUsageProvider
	disks     map[string]*diskUsage
}

type diskUsage struct {
	stats       filesystem.DiskStats
	lastUpdated time.Time
}

func NewDiskUsageCacher(fsHandler fsUsageProvider) *DiskUsageCacher {
	return &DiskUsageCacher{
		fsHandler: fsHandler,
		disks:     make(map[string]*diskUsage),
	}
}

func (c *DiskUsageCacher) GetDiskUsage(storage schema.Storage) (filesystem.DiskStats, error) {
	c.Lock()
	defer c.Unlock()

	cachedUsage, usageExistsInCache := c.disks[storage.GetName()]
	if usageExistsInCache {
		if time.Since(cachedUsage.lastUpdated) < time.Second {
			return cachedUsage.stats, nil
		}
	}

	statsNew, err := c.fsHandler.GetDiskUsage(storage.GetFSPath())
	if err != nil {
		if usageExistsInCache {
			// Allow stale stats to be used only once on error.
			cachedStats := c.disks[storage.GetName()].stats
			delete(c.disks, storage.GetName())

			return cachedStats, nil
		}

		return filesystem.DiskStats{}, fmt.Errorf("(alloc-usage) failed to get usage: %w", err)
	}

	if usageExistsInCache {
		c.disks[storage.GetName()].stats = statsNew
		c.disks[storage.GetName()].lastUpdated = time.Now()

		return statsNew, nil
	}

	c.disks[storage.GetName()] = &diskUsage{
		stats:       statsNew,
		lastUpdated: time.Now(),
	}

	return statsNew, nil
}

func (c *DiskUsageCacher) HasEnoughFreeSpace(s schema.Storage, minFree uint64, fileSize uint64) (bool, error) {
	stats, err := c.GetDiskUsage(s)
	if err != nil {
		return false, fmt.Errorf("(alloc-enoughfree) failed to get usage: %w", err)
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
