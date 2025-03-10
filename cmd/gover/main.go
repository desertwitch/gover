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

type taskHandlers struct {
	FSHandler     *filesystem.Handler
	UnraidHandler *unraid.Handler
	AllocHandler  *allocation.Handler
	IOHandler     *io.Handler
}

func processSystem(wg *sync.WaitGroup, ctx context.Context, system *unraid.System, deps *taskHandlers) {
	defer wg.Done()

	var qwg sync.WaitGroup

	files, err := getFilesBySystem(ctx, system, deps)
	if err != nil {
		slog.Error("Failure enumerating queue", "err", err)

		return
	}

	bqueue := queue.NewBucketQueue()
	bqueue.Enqueue(files...)

	bqueues := bqueue.GetQueuesUnsafe()

	maxWorkers := runtime.NumCPU()
	semaphore := make(chan struct{}, maxWorkers)

	osProvider := &filesystem.OS{}
	unixProvider := &filesystem.Unix{}
	ioOps := io.NewHandler(deps.AllocHandler, deps.FSHandler, osProvider, unixProvider)

	for _, q := range bqueues {
		qwg.Add(1)
		semaphore <- struct{}{}

		go func(q *queue.QueueManager) {
			defer qwg.Done()
			defer func() { <-semaphore }()

			if err := ioOps.ProcessQueue(ctx, q); err != nil {
				slog.Error("Failure processing a queue", "err", err)

				return
			}
		}(q)
	}
	qwg.Wait()
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
	unraidOps := unraid.NewHandler(fsOps, configOps)
	allocOps := allocation.NewHandler(fsOps)
	ioOps := io.NewHandler(allocOps, fsOps, osProvider, unixProvider)

	deps := &taskHandlers{
		FSHandler:     fsOps,
		UnraidHandler: unraidOps,
		AllocHandler:  allocOps,
		IOHandler:     ioOps,
	}

	system, err := unraidOps.EstablishSystem()
	if err != nil {
		slog.Error("Failed to establish (parts of) the Unraid system.",
			"err", err,
		)

		return
	}

	wg.Add(1)
	go processSystem(&wg, ctx, system, deps)
	wg.Wait()

	if ctx.Err() != nil {
		os.Exit(1)
	}
}
