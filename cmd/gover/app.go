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

func (app *App) LaunchUI() error {
	if err := app.uiHandler.Launch(); err != nil {
		return fmt.Errorf("(app-ui) %w", err)
	}

	return nil
}
