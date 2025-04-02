package main

import (
	"context"
	"fmt"

	"github.com/desertwitch/gover/internal/generic/allocation"
	"github.com/desertwitch/gover/internal/generic/filesystem"
	"github.com/desertwitch/gover/internal/generic/io"
	"github.com/desertwitch/gover/internal/generic/pathing"
	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/desertwitch/gover/internal/generic/schema"
	"github.com/desertwitch/gover/internal/generic/ui"
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

func (app *App) Launch(ctx context.Context) error {
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

func (app *App) LaunchUI(ctx context.Context, cancel context.CancelFunc) error {
	if err := app.uiHandler.Launch(ctx, cancel); err != nil {
		return fmt.Errorf("(app-ui) %w", err)
	}

	return nil
}
