package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
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
	ExitCode   = 0
	Version    string
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write memory profile to this file")
)

type App struct {
	shares         map[string]schema.Share
	fsHandler      *filesystem.Handler
	allocHandler   *allocation.Handler
	pathingHandler *pathing.Handler
	ioHandler      *io.Handler
	queueManager   *queue.Manager
	uiHandler      *ui.Handler
}

func NewApp(shares map[string]schema.Share,
	fsHandler *filesystem.Handler,
	allocHandler *allocation.Handler,
	pathingHandler *pathing.Handler,
	ioHandler *io.Handler,
	queueManager *queue.Manager,
	uiHandler *ui.Handler,
) *App {
	return &App{
		shares:         shares,
		fsHandler:      fsHandler,
		allocHandler:   allocHandler,
		pathingHandler: pathingHandler,
		ioHandler:      ioHandler,
		queueManager:   queueManager,
		uiHandler:      uiHandler,
	}
}

func (app *App) LaunchUI(ctx context.Context) error {
	if err := app.uiHandler.Launch(ctx); err != nil {
		return fmt.Errorf("(ui) %w", err)
	}

	return nil
}

func (app *App) Launch(ctx context.Context) error {
	if err := app.Enumerate(ctx); err != nil {
		return err
	}

	if err := app.Evaluate(ctx); err != nil {
		return err
	}

	if err := app.IO(ctx); err != nil {
		return err
	}

	return nil
}

//nolint:funlen
func main() {
	defer func() {
		os.Exit(ExitCode)
	}()

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	slog.Info("Warming up, a good day to move some files!", "version", Version)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	memObserver := NewMemoryObserver(ctx)
	defer memObserver.Stop()

	defer func() {
		slog.Info("Memory consumption peaked at:", "maxAlloc", (memObserver.GetMaxAlloc() / 1024 / 1024)) //nolint:mnd
	}()

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			slog.Error("Could not create cpu profile", "err", err)

			return
		}
		defer f.Close()

		_ = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *memprofile != "" {
		defer func() {
			f, err := os.Create(*memprofile)
			if err != nil {
				slog.Error("Could not create allocs profile", "err", err)
			}
			defer f.Close()

			if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
				slog.Error("Could not write allocs profile", "err", err)
			}
		}()
	}

	establishSignalHandlers(cancel)

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
	uiHandler := ui.NewHandler(queueManager)

	shareAdapters := make(map[string]schema.Share, len(shares))
	for name, share := range shares {
		shareAdapters[name] = NewShareAdapter(share)
	}

	app := NewApp(shareAdapters, fsHandler, allocHandler, pathingHandler, ioHandler, queueManager, uiHandler)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.Launch(ctx); err != nil {
			ExitCode = 1
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.LaunchUI(ctx); err != nil {
			slog.Error("UI failure: falling back to terminal.", "err", err)
		}
	}()

	wg.Wait()
}

func establishSignalHandlers(cancel context.CancelFunc) {
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
