package main

import (
	"context"
	"log/slog"
	"os"
	"runtime/pprof"
)

// cpuProfiler is an implementation for a [pprof] CPU profiler.
type cpuProfiler struct {
	ctx      context.Context //nolint:containedctx
	cancel   context.CancelFunc
	doneChan chan struct{}
}

// newCPUProfiler returns a pointer to a new [cpuProfiler]. The profiling is
// started, if a given path is not empty or nil. The profiling needs to be
// stopped by e.g. deferred calling [cpuProfiler.Stop] before program exit.
func newCPUProfiler(ctx context.Context, path *string) *cpuProfiler {
	cprof := &cpuProfiler{}
	cprof.ctx, cprof.cancel = context.WithCancel(ctx)
	cprof.doneChan = make(chan struct{})

	go cprof.profile(path)

	return cprof
}

// profile is the principal method for the CPU profiling.
func (cprof *cpuProfiler) profile(path *string) {
	defer close(cprof.doneChan)

	if path == nil || *path == "" {
		return
	}

	f, err := os.Create(*path)
	if err != nil {
		slog.Error("Could not create cpu profile", "err", err)

		return
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		slog.Error("Could not start cpu profile", "err", err)

		return
	}

	defer pprof.StopCPUProfile()

	<-cprof.ctx.Done()
}

// Stop signals for the CPU profiling to stop and be written out to the
// profiling file.
func (cprof *cpuProfiler) Stop() {
	cprof.cancel()
	<-cprof.doneChan
}

// allocProfiler is an implementation for a [pprof] memory profiler.
type allocProfiler struct {
	ctx      context.Context //nolint:containedctx
	cancel   context.CancelFunc
	doneChan chan struct{}
}

// newAllocProfiler returns a pointer to a new [allocProfiler]. The profiling is
// started, if a given path is not empty or nil. The profiling needs to be
// stopped by e.g. deferred calling [allocProfiler.Stop] before program exit.
func newAllocProfiler(ctx context.Context, path *string) *allocProfiler {
	aprof := &allocProfiler{}
	aprof.ctx, aprof.cancel = context.WithCancel(ctx)
	aprof.doneChan = make(chan struct{})

	go aprof.profile(path)

	return aprof
}

// profile is the principal method for the memory profiling.
func (aprof *allocProfiler) profile(path *string) {
	defer close(aprof.doneChan)

	if path == nil || *path == "" {
		return
	}

	<-aprof.ctx.Done()

	f, err := os.Create(*path)
	if err != nil {
		slog.Error("Could not create allocs profile", "err", err)
	}
	defer f.Close()

	if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
		slog.Error("Could not write allocs profile", "err", err)
	}
}

// Stop signals for the memory profiling to stop and be written out to the
// profiling file.
func (aprof *allocProfiler) Stop() {
	aprof.cancel()
	<-aprof.doneChan
}
