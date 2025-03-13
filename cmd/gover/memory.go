package main

import (
	"context"
	"runtime"
	"time"
)

const (
	memoryMonitorInterval = 100 * time.Millisecond
)

func memoryMonitor(ctx context.Context, ch chan<- uint64) {
	defer close(ch)

	var maxAlloc uint64

	ticker := time.NewTicker(memoryMonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ch <- maxAlloc

			return
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			if m.Alloc > maxAlloc {
				maxAlloc = m.Alloc
			}
		}
	}
}
