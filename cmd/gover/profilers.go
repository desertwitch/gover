package main

import (
	"context"
	"log/slog"
	"os"
	"runtime/pprof"
)

//nolint:containedctx
type cpuProfiler struct {
	ctx      context.Context
	cancel   context.CancelFunc
	doneChan chan struct{}
}

func newCPUProfiler(ctx context.Context, path *string) *cpuProfiler {
	cprof := &cpuProfiler{}
	cprof.ctx, cprof.cancel = context.WithCancel(ctx)
	cprof.doneChan = make(chan struct{})

	go cprof.Profile(path)

	return cprof
}

func (cprof *cpuProfiler) Profile(path *string) {
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

func (cprof *cpuProfiler) Stop() {
	cprof.cancel()
	<-cprof.doneChan
}

//nolint:containedctx
type allocProfiler struct {
	ctx      context.Context
	cancel   context.CancelFunc
	doneChan chan struct{}
}

func newAllocProfiler(ctx context.Context, path *string) *allocProfiler {
	aprof := &allocProfiler{}
	aprof.ctx, aprof.cancel = context.WithCancel(ctx)
	aprof.doneChan = make(chan struct{})

	go aprof.Profile(path)

	return aprof
}

func (aprof *allocProfiler) Profile(path *string) {
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

func (aprof *allocProfiler) Stop() {
	aprof.cancel()
	<-aprof.doneChan
}
