package main

import (
	"context"
	"log/slog"
	"os"
	"runtime/pprof"
)

//nolint:containedctx
type CPUProfiler struct {
	ctx      context.Context
	cancel   context.CancelFunc
	doneChan chan struct{}
}

func NewCPUProfiler(ctx context.Context, path *string) *CPUProfiler {
	cprof := &CPUProfiler{}
	cprof.ctx, cprof.cancel = context.WithCancel(ctx)
	cprof.doneChan = make(chan struct{})

	go cprof.Profile(path)

	return cprof
}

func (cprof *CPUProfiler) Profile(path *string) {
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

func (cprof *CPUProfiler) Stop() {
	cprof.cancel()
	<-cprof.doneChan
}

//nolint:containedctx
type AllocProfiler struct {
	ctx      context.Context
	cancel   context.CancelFunc
	doneChan chan struct{}
}

func NewAllocProfiler(ctx context.Context, path *string) *AllocProfiler {
	aprof := &AllocProfiler{}
	aprof.ctx, aprof.cancel = context.WithCancel(ctx)
	aprof.doneChan = make(chan struct{})

	go aprof.Profile(path)

	return aprof
}

func (aprof *AllocProfiler) Profile(path *string) {
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

func (aprof *AllocProfiler) Stop() {
	aprof.cancel()
	<-aprof.doneChan
}
