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
	// CheckerInterval is the interval at which the [InUseChecker] is updated.
	CheckerInterval = 5 * time.Second
)

// osReadsProvider defines methods needed to read a filesystem of the operating
// system.
type osReadsProvider interface {
	ReadDir(name string) ([]os.DirEntry, error)
	Readlink(name string) (string, error)
}

// InUseChecker caches paths which are currently in use by another process of
// the operating system. This allows for fast checks if a given path is in use,
// without overloading the OS with syscalls.
type InUseChecker struct {
	sync.RWMutex
	osHandler  osReadsProvider
	inUsePaths map[string]struct{}
	isUpdating atomic.Bool
}

// NewInUseChecker returns a pointer to a new [InUseChecker]. The update method
// is started, querying the OS for in-use paths every [CheckerInterval].
func NewInUseChecker(ctx context.Context, osHandler osProvider) (*InUseChecker, error) {
	checker := &InUseChecker{
		osHandler:  osHandler,
		inUsePaths: make(map[string]struct{}),
	}

	if err := checker.Update(); err != nil {
		return nil, err
	}

	go checker.periodicUpdate(ctx)

	return checker, nil
}

// IsInUse checks (the cache) if a path is currently in use by another process
// of the operating system.
func (c *InUseChecker) IsInUse(path string) bool {
	c.RLock()
	defer c.RUnlock()

	_, exists := c.inUsePaths[path]

	return exists
}

// periodicUpdate calls [InUseChecker.Update] every [CheckerInterval].
func (c *InUseChecker) periodicUpdate(ctx context.Context) {
	ticker := time.NewTicker(CheckerInterval)
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

// Update queries the operating system for all in-use paths and stores them in
// the [InUseChecker] cache. Since this is a time and resource intensive
// operation, this method is a no-op with an update in progress.
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
