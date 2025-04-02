package ui

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/desertwitch/gover/internal/generic/queue"
	"github.com/lmittmann/tint"
)

type Handler struct {
	queueManager *queue.Manager
	program      *tea.Program
	logHandler   *teaLogWriter

	modelReady atomic.Bool

	Ready  atomic.Bool
	Failed atomic.Bool
}

func NewHandler(queueManager *queue.Manager) *Handler {
	return &Handler{
		queueManager: queueManager,
	}
}

func (uiHandler *Handler) setupLogging() {
	slog.SetDefault(slog.New(
		tint.NewHandler(uiHandler.logHandler, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
}

func (uiHandler *Handler) Launch(ctx context.Context, cancel context.CancelFunc) error {
	model := newTeaModel(uiHandler, uiHandler.queueManager, cancel)
	uiHandler.program = tea.NewProgram(model, tea.WithAltScreen(), tea.WithContext(ctx))

	uiHandler.logHandler = newTeaLogWriter(uiHandler.program)
	defer uiHandler.logHandler.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if uiHandler.Failed.Load() {
				return
			}

			if uiHandler.modelReady.Load() {
				uiHandler.setupLogging()
				uiHandler.Ready.Store(true)

				return
			}
		}
	}()

	if _, err := uiHandler.program.Run(); err != nil {
		uiHandler.Failed.Store(true)

		return fmt.Errorf("(ui-tea) %w", err)
	}

	return nil
}
