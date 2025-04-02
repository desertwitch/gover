package main

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

const (
	memoryMonitorInterval = 100 * time.Millisecond
)

type memoryObserver struct {
	sync.RWMutex
	maxAlloc uint64
	stopChan chan struct{}
}

func newMemoryObserver(ctx context.Context) *memoryObserver {
	obs := &memoryObserver{
		stopChan: make(chan struct{}),
	}
	go obs.monitor(ctx)

	return obs
}

func (o *memoryObserver) GetMaxAlloc() uint64 {
	o.RLock()
	defer o.RUnlock()

	return o.maxAlloc
}

func (o *memoryObserver) Stop() {
	close(o.stopChan)
	slog.Info("Memory consumption peaked at:", "maxAlloc", (o.GetMaxAlloc() / 1024 / 1024)) //nolint:mnd
}

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
