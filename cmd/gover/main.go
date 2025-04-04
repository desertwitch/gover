package main

import (
	"context"
	"errors"
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

	slogMan = NewSlogManager()

	uiEnabled  = flag.Bool("ui", true, "enable the UI")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write memory profile to this file")
)

func termLogging(enabled bool) {
	if enabled {
		if _, ok := slogMan.GetHandler("term"); !ok {
			slogMan.AddHandler("term",
				tint.NewHandler(os.Stdout,
					&tint.Options{
						Level:      slog.LevelDebug,
						TimeFormat: time.Kitchen,
					}),
			)
		}
	} else {
		slogMan.RemoveHandler("term")
	}
}

func uiLogging(enabled bool, writer *ui.TeaLogWriter) {
	if enabled {
		if _, ok := slogMan.GetHandler("ui"); !ok {
			slogMan.AddHandler("ui",
				tint.NewHandler(writer,
					&tint.Options{
						Level:      slog.LevelDebug,
						TimeFormat: time.Kitchen,
					}),
			)
		}
	} else {
		slogMan.RemoveHandler("ui")
	}
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

func startUI(ctx context.Context, wg *sync.WaitGroup, app *App) {
	defer wg.Done()

	if app.uiHandler != nil {
		var err error

		defer func() {
			uiLogging(false, nil)
			termLogging(true)

			if err != nil && !errors.Is(err, context.Canceled) {
				slog.Error("UI failure, falling back to regular terminal.",
					"err", err,
				)
			}
		}()

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if app.uiHandler.Failed.Load() {
					return
				}

				if app.uiHandler.Ready.Load() {
					termLogging(false)
					uiLogging(true, app.uiHandler.LogWriter)

					return
				}
			}
		}()

		err = app.LaunchUI()
	}
}

func main() {
	defer func() {
		os.Exit(ExitCode)
	}()

	slog.SetDefault(slog.New(slogMan))
	termLogging(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flag.Parse()
	setupSignalHandlers(cancel)

	memObserver := NewMemoryObserver(ctx)
	defer memObserver.Stop()

	cpuProfiler := NewCPUProfiler(ctx, cpuprofile)
	defer cpuProfiler.Stop()

	allocProfiler := NewAllocProfiler(ctx, memprofile)
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
		uiHandler = ui.NewHandler(ctx, cancel, queueManager)
	}

	shareAdapters := make(map[string]schema.Share, len(shares))
	for name, share := range shares {
		shareAdapters[name] = NewShareAdapter(share)
	}

	var wg sync.WaitGroup
	app := NewApp(shareAdapters, fsHandler, allocHandler, pathingHandler, ioHandler, queueManager, uiHandler)

	wg.Add(1)
	go startUI(ctx, &wg, app)

	wg.Add(1)
	go startApp(ctx, &wg, app)

	wg.Wait()
}
