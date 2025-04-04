package main

import (
	"context"
	"log/slog"
	"sync"
)

type SlogManager struct {
	sync.RWMutex
	handlers map[string]slog.Handler
	attrs    []slog.Attr
	groups   []string
}

func NewSlogManager() *SlogManager {
	return &SlogManager{
		handlers: make(map[string]slog.Handler),
	}
}

func (m *SlogManager) Enabled(ctx context.Context, level slog.Level) bool {
	m.RLock()
	defer m.RUnlock()

	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (m *SlogManager) Handle(ctx context.Context, r slog.Record) error {
	m.RLock()
	defer m.RUnlock()

	for _, h := range m.handlers {
		_ = h.Handle(ctx, r)
	}

	return nil
}

func (m *SlogManager) WithAttrs(attrs []slog.Attr) slog.Handler {
	m.Lock()
	defer m.Unlock()

	groups := make([]string, len(m.groups))
	copy(groups, m.groups)

	newLm := &SlogManager{
		handlers: make(map[string]slog.Handler, len(m.handlers)),
		attrs:    append(m.attrs, attrs...),
		groups:   groups,
	}

	for name, h := range m.handlers {
		newLm.handlers[name] = h.WithAttrs(attrs)
	}

	return newLm
}

func (m *SlogManager) WithGroup(name string) slog.Handler {
	m.Lock()
	defer m.Unlock()

	attrs := make([]slog.Attr, len(m.attrs))
	copy(attrs, m.attrs)

	newLm := &SlogManager{
		handlers: make(map[string]slog.Handler, len(m.handlers)),
		attrs:    attrs,
		groups:   append(m.groups, name),
	}

	for handlerName, h := range m.handlers {
		newLm.handlers[handlerName] = h.WithGroup(name)
	}

	return newLm
}

//nolint:unparam
func (m *SlogManager) GetHandler(name string) (slog.Handler, bool) {
	m.RLock()
	defer m.RUnlock()

	h, ok := m.handlers[name]

	return h, ok
}

func (m *SlogManager) AddHandler(name string, handler slog.Handler) {
	m.Lock()
	defer m.Unlock()

	h := handler
	for _, attr := range m.attrs {
		h = h.WithAttrs([]slog.Attr{attr})
	}

	for _, group := range m.groups {
		h = h.WithGroup(group)
	}

	m.handlers[name] = h
}

func (m *SlogManager) RemoveHandler(name string) {
	m.Lock()
	defer m.Unlock()

	delete(m.handlers, name)
}
