package main

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

const (
	// memoryMonitorInterval is the interval at which a [memoryObserver] is updated.
	memoryMonitorInterval = 100 * time.Millisecond
)

// memoryObserver tracks peak memory usage over a period of time.
type memoryObserver struct {
	sync.RWMutex
	maxAlloc uint64
	stopChan chan struct{}
}

// newMemoryObserver returns a pointer to a new [memoryObserver]. The tracking
// is started and needs to be stopped by e.g. deferred calling of
// [memoryObserver.Stop] before program exit.
func newMemoryObserver(ctx context.Context) *memoryObserver {
	obs := &memoryObserver{
		stopChan: make(chan struct{}),
	}
	go obs.monitor(ctx)

	return obs
}

// GetMaxAlloc returns the peak recorded memory allocation size in a
// thread-safe manner.
func (o *memoryObserver) GetMaxAlloc() uint64 {
	o.RLock()
	defer o.RUnlock()

	return o.maxAlloc
}

// Stops halts the tracking of peak memory allocation and logs the highest
// recorded memory allocation with [slog.Info]. It is usually called at the end
// of a program's lifetime.
func (o *memoryObserver) Stop() {
	close(o.stopChan)
	slog.Info("Memory consumption peaked at:", "maxAlloc", (o.GetMaxAlloc() / 1024 / 1024)) //nolint:mnd
}

// monitor is the principal method that queries the [runtime.MemStats] every
// [memoryMonitorInterval].
func (o *memoryObserver) monitor(ctx context.Context) {
	ticker := time.NewTicker(memoryMonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-o.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			o.Lock()
			if m.Alloc > o.maxAlloc {
				o.maxAlloc = m.Alloc
			}
			o.Unlock()
		}
	}
}
