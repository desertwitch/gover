package main

import (
	"context"
	"flag"
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
	"github.com/desertwitch/gover/internal/generic/schema"
	"github.com/desertwitch/gover/internal/generic/ui"
	"github.com/desertwitch/gover/internal/unraid"
	"github.com/lmittmann/tint"
)

const (
	stackTraceBufMax = 1 << 24
)

//nolint:gochecknoglobals
var (
	ExitCode = 0
	Version  string

	uiEnabled  = flag.Bool("ui", true, "enable the UI")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write memory profile to this file")
)

func setupLogging() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
}

func setupSignalHandlers(cancel context.CancelFunc) {
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

	sigChan3 := make(chan os.Signal, 1)
	signal.Notify(sigChan3, syscall.SIGUSR2)
	go func() {
		for range sigChan3 {
			runtime.GC()
		}
	}()
}

func startApp(ctx context.Context, wg *sync.WaitGroup, app *App) {
	defer wg.Done()

	if app.uiHandler != nil {
		slog.Info("Waiting for UI...")
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if app.uiHandler.Ready.Load() || app.uiHandler.Failed.Load() {
				break
			}
		}
	}

	if err := app.Launch(ctx); err != nil {
		ExitCode = 1
	}
}

func startUI(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, app *App) {
	defer wg.Done()

	if app.uiHandler != nil {
		defer setupLogging()

		if err := app.LaunchUI(ctx, cancel); err != nil {
			slog.Error("UI failure: falling back to terminal.", "err", err)
		}
	}
}

func main() {
	defer func() {
		os.Exit(ExitCode)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flag.Parse()
	setupLogging()
	setupSignalHandlers(cancel)

	memObserver := newMemoryObserver(ctx)
	defer memObserver.Stop()

	cpuProfiler := newCPUProfiler(ctx, cpuprofile)
	defer cpuProfiler.Stop()

	allocProfiler := newAllocProfiler(ctx, memprofile)
	defer allocProfiler.Stop()

	osProvider := &schema.OS{}
	unixProvider := &schema.Unix{}
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
	unraidHandler := unraid.NewHandler(fsHandler, configHandler, osProvider)

	system, err := unraidHandler.EstablishSystem()
	if err != nil {
		slog.Error("Failed to establish (parts of) the Unraid system.",
			"err", err,
		)

		return
	}

	shares := system.GetShares()
	queueManager := queue.NewManager()

	var uiHandler *ui.Handler
	if uiEnabled != nil && *uiEnabled {
		uiHandler = ui.NewHandler(queueManager)
	}

	shareAdapters := make(map[string]schema.Share, len(shares))
	for name, share := range shares {
		shareAdapters[name] = NewShareAdapter(share)
	}

	var wg sync.WaitGroup
	app := NewApp(shareAdapters, fsHandler, allocHandler, pathingHandler, ioHandler, queueManager, uiHandler)

	wg.Add(1)
	go startUI(ctx, cancel, &wg, app)

	wg.Add(1)
	go startApp(ctx, &wg, app)

	wg.Wait()
}
