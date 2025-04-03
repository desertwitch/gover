package ui

import (
	"context"
	"fmt"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/desertwitch/gover/internal/generic/queue"
)

type Handler struct {
	queueManager *queue.Manager
	program      *tea.Program

	LogWriter *TeaLogWriter

	Ready  atomic.Bool
	Failed atomic.Bool
}

func NewHandler(ctx context.Context, cancel context.CancelFunc, queueManager *queue.Manager) *Handler {
	handler := &Handler{
		queueManager: queueManager,
	}

	model := newTeaModel(handler, queueManager, cancel)
	handler.program = tea.NewProgram(model, tea.WithAltScreen(), tea.WithContext(ctx))
	handler.LogWriter = NewTeaLogWriter(handler.program)

	return handler
}

func (uiHandler *Handler) Launch() error {
	defer uiHandler.LogWriter.Stop()

	if _, err := uiHandler.program.Run(); err != nil {
		uiHandler.Failed.Store(true)

		return fmt.Errorf("(ui) %w", err)
	}

	return nil
}
