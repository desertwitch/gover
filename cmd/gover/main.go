/*
gover is a mover-type application for moving files between various storage, and
according to a defined set of rules, configuration and logical pathways.
*/
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

	"github.com/desertwitch/gover/internal/allocation"
	"github.com/desertwitch/gover/internal/configuration"
	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/io"
	"github.com/desertwitch/gover/internal/pathing"
	"github.com/desertwitch/gover/internal/queue"
	"github.com/desertwitch/gover/internal/schema"
	"github.com/desertwitch/gover/internal/ui"
	"github.com/desertwitch/gover/internal/unraid"
	"github.com/lmittmann/tint"
)

const (
	// stackTraceBufMax is the limiting size for a requested stack trace.
	stackTraceBufMax = 1 << 24
)

var (
	// Version is the application's version (filled in during compilation).
	Version string

	exitCode = 0
	slogMan  = newSlogManager()

	uiEnabled = flag.Bool("ui", true, "enable the UI")
)

// termLogging enables or disables logs to be sent to the terminal (via
// [os.Stdout]).
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

// uiLogging enables or disables logs to be sent to a user interface (via
// [ui.TeaLogWriter]).
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

// setupSignalHandlers setups up operating system singal handling.
//   - SIGTERM, SIGINT initiate graceful program teardown.
//   - SIGUSR1 initiates printing of a stack trace.
//   - SIGUSR2 initiates forced garbage collection.
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
			_, _ = os.Stderr.Write(buf[:stacklen])
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

// startApp is a helper function to start the application, waiting for the user
// interface to come up or fail (if one was requested for the application).
func startApp(ctx context.Context, wg *sync.WaitGroup, app *app) {
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
		exitCode = 1
	}
}

// startUI is a helper function to start the application's user interface. If no
// user interface was requested for the application, this function is a no-op.
func startUI(ctx context.Context, wg *sync.WaitGroup, app *app) {
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
		os.Exit(exitCode)
	}()

	slog.SetDefault(slog.New(slogMan))
	termLogging(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flag.Parse()
	setupSignalHandlers(cancel)

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
	pathingHandler := pathing.NewHandler(osProvider)
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
		shareAdapters[name] = newShareAdapter(share)
	}

	var wg sync.WaitGroup
	app := newApp(shareAdapters, queueManager, fsHandler, allocHandler, pathingHandler, ioHandler, uiHandler)

	wg.Add(1)
	go startUI(ctx, &wg, app)

	wg.Add(1)
	go startApp(ctx, &wg, app)

	wg.Wait()
}
