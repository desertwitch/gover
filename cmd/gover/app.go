package main

import (
	"context"
	"fmt"

	"github.com/desertwitch/gover/internal/allocation"
	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/io"
	"github.com/desertwitch/gover/internal/pathing"
	"github.com/desertwitch/gover/internal/queue"
	"github.com/desertwitch/gover/internal/schema"
	"github.com/desertwitch/gover/internal/ui"
)

// app is the principal implementation of the application.
type app struct {
	// shares is a map of all [schema.Share] that will be checked.
	shares map[string]schema.Share // map[shareName]schema.Share

	// queueManager is a [queue.Manager] for all application operations.
	queueManager *queue.Manager

	fsHandler      *filesystem.Handler
	allocHandler   *allocation.Handler
	pathingHandler *pathing.Handler
	ioHandler      *io.Handler
	uiHandler      *ui.Handler
}

// newApp returns a pointer to a new [app].
func newApp(shares map[string]schema.Share,
	queueManager *queue.Manager,
	fsHandler *filesystem.Handler,
	allocHandler *allocation.Handler,
	pathingHandler *pathing.Handler,
	ioHandler *io.Handler,
	uiHandler *ui.Handler,
) *app {
	return &app{
		shares:         shares,
		queueManager:   queueManager,
		fsHandler:      fsHandler,
		allocHandler:   allocHandler,
		pathingHandler: pathingHandler,
		ioHandler:      ioHandler,
		uiHandler:      uiHandler,
	}
}

// Launch starts the application and it's subtasks:
//   - Enumeration to collect all [schema.Moveable] candidates.
//   - Evaluation to sort, allocate and validate all [schema.Moveable].
//   - IO to move all [schema.Moveable] to their final destinations.
func (app *app) Launch(ctx context.Context) error {
	if err := app.Enumerate(ctx); err != nil {
		return fmt.Errorf("(app) %w", err)
	}

	if err := app.Evaluate(ctx); err != nil {
		return fmt.Errorf("(app) %w", err)
	}

	if err := app.IO(ctx); err != nil {
		return fmt.Errorf("(app) %w", err)
	}

	return nil
}

// LaunchUI starts the application's command-line user interface.
func (app *app) LaunchUI() error {
	if err := app.uiHandler.Launch(); err != nil {
		return fmt.Errorf("(app-ui) %w", err)
	}

	return nil
}
