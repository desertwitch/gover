package ui

import (
	"context"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/desertwitch/gover/internal/generic/queue"
)

type Handler struct {
	queueManager *queue.Manager
	logHandler   *TeaLogger
	program      *tea.Program
}

func NewHandler(queueManager *queue.Manager) *Handler {
	return &Handler{
		queueManager: queueManager,
		logHandler:   NewLogHandler(),
	}
}

func (uiHandler *Handler) Launch(ctx context.Context) error {
	// Set up the slog handler
	logger := slog.New(uiHandler.logHandler)
	slog.SetDefault(logger)

	// Create a new tea program
	model := NewTeaModel(uiHandler.queueManager, uiHandler.logHandler)

	uiHandler.program = tea.NewProgram(model, tea.WithAltScreen(), tea.WithContext(ctx))

	// Set program reference in the log handler to enable message passing
	uiHandler.logHandler.SetProgram(uiHandler.program)

	// Start the program
	if _, err := uiHandler.program.Run(); err != nil {
		return fmt.Errorf("(ui-tea) %w", err)
	}

	return nil
}

func (uiHandler *Handler) Stop() {
	uiHandler.program.Kill()
}
