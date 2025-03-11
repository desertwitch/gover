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

	"github.com/desertwitch/gover/internal/adapters/unraid"
	"github.com/desertwitch/gover/internal/generic/allocation"
	"github.com/desertwitch/gover/internal/generic/configuration"
	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/io"
	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/storage"
	"github.com/lmittmann/tint"
)

type depPackage struct {
	FSHandler    *filesystem.Handler
	AllocHandler *allocation.Handler
	IOHandler    *io.Handler
}

func newDepPackage(fsHandler *filesystem.Handler, allocHandler *allocation.Handler, ioHandler *io.Handler) *depPackage {
	return &depPackage{
		FSHandler:    fsHandler,
		AllocHandler: allocHandler,
		IOHandler:    ioHandler,
	}
}

func processSystem(ctx context.Context, wg *sync.WaitGroup, system storage.System, deps *depPackage) {
	defer wg.Done()

	queueMan := queue.NewManager()

	files, err := enumerateShares(ctx, system.GetShares(), system.GetArray().GetDisks(), deps)
	if err != nil {
		return
	}

	queueMan.Enqueue(files...)
	destQueues := queueMan.GetQueuesUnsafe()

	var queueWG sync.WaitGroup
	maxWorkers := runtime.NumCPU()
	semaphore := make(chan struct{}, maxWorkers)

	for _, destQueue := range destQueues {
		semaphore <- struct{}{}

		queueWG.Add(1)
		go func(q *queue.DestinationQueue) {
			defer queueWG.Done()
			defer func() { <-semaphore }()

			deps.IOHandler.ProcessQueue(ctx, q)
		}(destQueue)
	}

	queueWG.Wait()
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		cancel()
	}()

	osProvider := &filesystem.OS{}
	unixProvider := &filesystem.Unix{}
	configProvider := &configuration.GodotenvProvider{}

	configHandler := configuration.NewHandler(configProvider)
	fsHandler := filesystem.NewHandler(osProvider, unixProvider)
	allocHandler := allocation.NewHandler(fsHandler)
	ioHandler := io.NewHandler(fsHandler, osProvider, unixProvider)

	unraidHandler := unraid.NewHandler(fsHandler, configHandler)

	system, err := unraidHandler.EstablishSystem()
	if err != nil {
		slog.Error("Failed to establish (parts of) the Unraid system.",
			"err", err,
		)

		return
	}

	deps := newDepPackage(fsHandler, allocHandler, ioHandler)

	wg.Add(1)
	go processSystem(ctx, &wg, system, deps)
	wg.Wait()

	if ctx.Err() != nil {
		os.Exit(1)
	}
}
