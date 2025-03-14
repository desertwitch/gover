package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/desertwitch/gover/internal/generic/allocation"
	"github.com/desertwitch/gover/internal/generic/configuration"
	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/io"
	"github.com/desertwitch/gover/internal/generic/pathing"
	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/storage"
	"github.com/desertwitch/gover/internal/unraid"
	"github.com/lmittmann/tint"
)

const (
	stackTraceBufMax = 1 << 24
)

type depPackage struct {
	FSHandler      *filesystem.Handler
	AllocHandler   *allocation.Handler
	PathingHandler *pathing.Handler
	IOHandler      *io.Handler
}

func newDepPackage(fsHandler *filesystem.Handler, allocHandler *allocation.Handler, pathingHandler *pathing.Handler, ioHandler *io.Handler) *depPackage {
	return &depPackage{
		FSHandler:      fsHandler,
		AllocHandler:   allocHandler,
		PathingHandler: pathingHandler,
		IOHandler:      ioHandler,
	}
}

func processShares(ctx context.Context, wg *sync.WaitGroup, shares map[string]storage.Share, queueMan *queue.Manager, deps *depPackage) {
	defer wg.Done()

	files, err := enumerateShares(ctx, shares, queueMan, deps)
	if err != nil {
		return
	}

	if err := ioProcessFiles(ctx, files, queueMan, deps); err != nil {
		return
	}
}

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		cancel()
	}()

	sigChan2 := make(chan os.Signal, 1)
	signal.Notify(sigChan2, syscall.SIGUSR1)
	go func() {
		for range sigChan2 {
			buf := make([]byte, stackTraceBufMax)
			stacklen := runtime.Stack(buf, true)
			os.Stderr.Write(buf[:stacklen])
		}
	}()

	memChan := make(chan uint64, 1)
	go memoryMonitor(ctx, memChan)

	osProvider := &filesystem.OS{}
	unixProvider := &filesystem.Unix{}
	configProvider := &configuration.GodotenvProvider{}

	fsHandler, err := filesystem.NewHandler(ctx, osProvider, unixProvider)
	if err != nil {
		slog.Error("Failed to establish filesystem handler.",
			"err", err,
		)

		return
	}

	allocHandler := allocation.NewHandler(fsHandler)
	pathingHandler := pathing.NewHandler(fsHandler)
	ioHandler := io.NewHandler(fsHandler, osProvider, unixProvider)

	configHandler := configuration.NewHandler(configProvider)
	unraidHandler := unraid.NewHandler(fsHandler, configHandler)

	system, err := unraidHandler.EstablishSystem()
	if err != nil {
		slog.Error("Failed to establish (parts of) the Unraid system.",
			"err", err,
		)

		return
	}

	deps := newDepPackage(fsHandler, allocHandler, pathingHandler, ioHandler)

	shares := system.GetShares()
	queueMan := queue.NewManager()

	shareAdapters := make(map[string]storage.Share, len(shares))
	for name, share := range shares {
		shareAdapters[name] = NewShareAdapter(share)
	}

	wg.Add(1)
	go processShares(ctx, &wg, shareAdapters, queueMan, deps)
	wg.Wait()

	cancel()

	slog.Info("Memory consumption peaked at:", "maxAlloc", (<-memChan / 1024 / 1024))

	if ctx.Err() != nil {
		os.Exit(1)
	}
}
