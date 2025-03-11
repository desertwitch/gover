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

	"github.com/desertwitch/gover/internal/allocation"
	"github.com/desertwitch/gover/internal/configuration"
	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/io"
	"github.com/desertwitch/gover/internal/queue"
	"github.com/desertwitch/gover/internal/unraid"
	"github.com/lmittmann/tint"
)

type depPackage struct {
	FSHandler    *filesystem.Handler
	AllocHandler *allocation.Handler
	IOHandler    *io.Handler
}

func newDepPackage(fsOps *filesystem.Handler, allocOps *allocation.Handler, ioOps *io.Handler) *depPackage {
	return &depPackage{
		FSHandler:    fsOps,
		AllocHandler: allocOps,
		IOHandler:    ioOps,
	}
}

func processSystem(ctx context.Context, wg *sync.WaitGroup, system filesystem.SystemType, deps *depPackage) {
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

	configOps := configuration.NewHandler(configProvider)
	fsOps := filesystem.NewHandler(osProvider, unixProvider)
	allocOps := allocation.NewHandler(fsOps)
	ioOps := io.NewHandler(fsOps, osProvider, unixProvider)

	unraidOps := unraid.NewHandler(fsOps, configOps)

	system, err := unraidOps.EstablishSystem()
	if err != nil {
		slog.Error("Failed to establish (parts of) the Unraid system.",
			"err", err,
		)

		return
	}

	deps := newDepPackage(fsOps, allocOps, ioOps)

	wg.Add(1)
	go processSystem(ctx, &wg, system, deps)
	wg.Wait()

	if ctx.Err() != nil {
		os.Exit(1)
	}
}
