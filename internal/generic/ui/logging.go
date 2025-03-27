package ui

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

type TeaLogger struct {
	sync.RWMutex
	program *tea.Program
}

func NewLogHandler() *TeaLogger {
	return &TeaLogger{}
}

func (h *TeaLogger) SetProgram(p *tea.Program) {
	h.Lock()
	defer h.Unlock()

	h.program = p
}

func (h *TeaLogger) Handle(ctx context.Context, r slog.Record) error {
	timeStr := r.Time.Format("15:04:05")
	levelStr := r.Level.String()

	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))

		return true
	})

	attrStr := ""
	if len(attrs) > 0 {
		attrStr = " " + fmt.Sprint(attrs)
	}

	logMsg := fmt.Sprintf("[%s] [%s] %s%s", timeStr, levelStr, r.Message, attrStr)

	h.RLock()
	if h.program != nil {
		h.program.Send(logEntryMsg(logMsg))
	}
	h.RUnlock()

	return nil
}

func (h *TeaLogger) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *TeaLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *TeaLogger) WithGroup(name string) slog.Handler {
	return h
}
