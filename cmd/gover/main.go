package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/desertwitch/gover/internal/generic/allocation"
	"github.com/desertwitch/gover/internal/generic/configuration"
	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/io"
	"github.com/desertwitch/gover/internal/generic/storage"
	"github.com/desertwitch/gover/internal/unraid"
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
	shares := system.GetShares()

	shareAdapters := make(map[string]storage.Share, len(shares))
	for name, share := range shares {
		shareAdapters[name] = NewShareAdapter(share)
	}

	wg.Add(1)
	go processShares(ctx, &wg, shareAdapters, deps)
	wg.Wait()

	if ctx.Err() != nil {
		os.Exit(1)
	}
}
