package main

import (
	"context"
	"runtime"
	"sync"
	"time"
)

const (
	memoryMonitorInterval = 100 * time.Millisecond
)

type MemoryObserver struct {
	sync.RWMutex
	maxAlloc uint64
	stopChan chan struct{}
}

func NewMemoryObserver(ctx context.Context) *MemoryObserver {
	obs := &MemoryObserver{
		stopChan: make(chan struct{}),
	}
	go obs.monitor(ctx)

	return obs
}

func (o *MemoryObserver) GetMaxAlloc() uint64 {
	o.RLock()
	defer o.RUnlock()

	return o.maxAlloc
}

func (o *MemoryObserver) Stop() {
	close(o.stopChan)
}

func (o *MemoryObserver) monitor(ctx context.Context) {
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
