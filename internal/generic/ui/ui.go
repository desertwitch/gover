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

//nolint:containedctx
type Handler struct {
	queueManager *queue.Manager
	program      *tea.Program
	logHandler   *teaLogWriter

	ctx        context.Context
	modelReady atomic.Bool

	Ready  atomic.Bool
	Failed atomic.Bool
}

func NewHandler(ctx context.Context, cancel context.CancelFunc, queueManager *queue.Manager) *Handler {
	handler := &Handler{
		queueManager: queueManager,
		ctx:          ctx,
	}

	model := newTeaModel(handler, queueManager, cancel)
	handler.program = tea.NewProgram(model, tea.WithAltScreen(), tea.WithContext(ctx))
	handler.logHandler = newTeaLogWriter(handler.program)

	return handler
}

func (uiHandler *Handler) setupLogging() {
	slog.SetDefault(slog.New(
		tint.NewHandler(uiHandler.logHandler, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
}

func (uiHandler *Handler) Launch() error {
	defer uiHandler.logHandler.Stop()

	go func() {
		for {
			select {
			case <-uiHandler.ctx.Done():
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

		return fmt.Errorf("(ui) %w", err)
	}

	return nil
}
