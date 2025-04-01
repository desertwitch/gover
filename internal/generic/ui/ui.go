package ui

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/lmittmann/tint"
)

type Handler struct {
	queueManager *queue.Manager
	logHandler   *teaLogWriter
	program      *tea.Program
}

func NewHandler(queueManager *queue.Manager) *Handler {
	return &Handler{
		queueManager: queueManager,
		logHandler:   newTeaLogWriter(),
	}
}

func (uiHandler *Handler) Launch(ctx context.Context, cancel context.CancelFunc) error {
	// Create a new tea program
	model := NewTeaModel(uiHandler.queueManager, uiHandler.logHandler, cancel)

	uiHandler.program = tea.NewProgram(model, tea.WithAltScreen(), tea.WithContext(ctx))

	// Redirect logs to UI
	uiHandler.logHandler.SetProgram(uiHandler.program)

	uiHandler.logHandler.Start()
	defer uiHandler.logHandler.Stop()

	slog.SetDefault(slog.New(
		tint.NewHandler(uiHandler.logHandler, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	// Start the program
	if _, err := uiHandler.program.Run(); err != nil {
		return fmt.Errorf("(ui-tea) %w", err)
	}

	return nil
}

func (uiHandler *Handler) Stop() {
	uiHandler.program.Kill()
}
