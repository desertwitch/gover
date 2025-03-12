package filesystem

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	CheckerInterval = 5 * time.Second
)

type osReadsProvider interface {
	ReadDir(name string) ([]os.DirEntry, error)
	Readlink(name string) (string, error)
}

type InUseChecker struct {
	sync.RWMutex
	osHandler   osReadsProvider
	inUsePaths  map[string]struct{}
	isUpdating  atomic.Bool
	stopUpdates chan struct{}
}

func NewInUseChecker(ctx context.Context, osHandler osProvider) (*InUseChecker, error) {
	checker := &InUseChecker{
		osHandler:   osHandler,
		inUsePaths:  make(map[string]struct{}),
		stopUpdates: make(chan struct{}),
	}

	if err := checker.Update(); err != nil {
		return nil, err
	}

	go checker.periodicUpdate(ctx)

	return checker, nil
}

func (f *Handler) IsInUse(path string) bool {
	return f.inUseHandler.IsInUse(path)
}

func (c *InUseChecker) IsInUse(path string) bool {
	c.RLock()
	defer c.RUnlock()

	_, exists := c.inUsePaths[path]

	return exists
}

func (c *InUseChecker) Update() error {
	if !c.isUpdating.CompareAndSwap(false, true) {
		return nil
	}

	c.Lock()
	defer func() {
		c.Unlock()
		c.isUpdating.Store(false)
	}()

	c.inUsePaths = make(map[string]struct{})

	procEntries, err := c.osHandler.ReadDir("/proc")
	if err != nil {
		return fmt.Errorf("failed to read /proc: %w", err)
	}

	for _, procEntry := range procEntries {
		pid, err := strconv.Atoi(procEntry.Name())
		if err != nil {
			continue
		}

		fdPath := fmt.Sprintf("/proc/%d/fd", pid)
		fdEntries, err := c.osHandler.ReadDir(fdPath)
		if err != nil {
			continue
		}

		for _, fdEntry := range fdEntries {
			fdLink := fmt.Sprintf("/proc/%d/fd/%s", pid, fdEntry.Name())

			linkTarget, err := c.osHandler.Readlink(fdLink)
			if err != nil {
				continue
			}

			c.inUsePaths[linkTarget] = struct{}{}
		}
	}

	return nil
}

func (c *InUseChecker) Stop() {
	close(c.stopUpdates)
}

func (c *InUseChecker) periodicUpdate(ctx context.Context) {
	ticker := time.NewTicker(CheckerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopUpdates:
			return
		case <-ticker.C:
			_ = c.Update()
		}
	}
}
