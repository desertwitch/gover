package main

import (
	"context"
	"log/slog"
	"sync"
)

// slogManager is an implementation of a [slog.Handler] that is able to manage
// multiple [slog.Handler] children for logging to multiple targets.
type slogManager struct {
	sync.RWMutex
	handlers map[string]slog.Handler
	attrs    []slog.Attr
	groups   []string
}

// newSlogManager returns a pointer to a new [slogManager].
func newSlogManager() *slogManager {
	return &slogManager{
		handlers: make(map[string]slog.Handler),
	}
}

// Enabled calls the Enabled method on each managed [slog.Handler].
func (m *slogManager) Enabled(ctx context.Context, level slog.Level) bool {
	m.RLock()
	defer m.RUnlock()

	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

// Handle calls the Handle method on each managed [slog.Handler].
func (m *slogManager) Handle(ctx context.Context, r slog.Record) error {
	m.RLock()
	defer m.RUnlock()

	for _, h := range m.handlers {
		_ = h.Handle(ctx, r)
	}

	return nil
}

// WithAttrs returns a new [slogManager] with the given attributes set on each
// of its managed [slog.Handler].
func (m *slogManager) WithAttrs(attrs []slog.Attr) slog.Handler {
	m.Lock()
	defer m.Unlock()

	groups := make([]string, len(m.groups))
	copy(groups, m.groups)

	newLm := &slogManager{
		handlers: make(map[string]slog.Handler, len(m.handlers)),
		attrs:    append(m.attrs, attrs...),
		groups:   groups,
	}

	for name, h := range m.handlers {
		newLm.handlers[name] = h.WithAttrs(attrs)
	}

	return newLm
}

// WithGroup returns a new [slogManager] with the given group set on each of its
// managed [slog.Handler].
func (m *slogManager) WithGroup(name string) slog.Handler {
	m.Lock()
	defer m.Unlock()

	attrs := make([]slog.Attr, len(m.attrs))
	copy(attrs, m.attrs)

	newLm := &slogManager{
		handlers: make(map[string]slog.Handler, len(m.handlers)),
		attrs:    attrs,
		groups:   append(m.groups, name),
	}

	for handlerName, h := range m.handlers {
		newLm.handlers[handlerName] = h.WithGroup(name)
	}

	return newLm
}

// GetHandler returns a specific managed [slog.Handler].
func (m *slogManager) GetHandler(name string) (slog.Handler, bool) { //nolint:unparam
	m.RLock()
	defer m.RUnlock()

	h, ok := m.handlers[name]

	return h, ok
}

// AddHandler adds a new [slog.Handler] to be managed.
func (m *slogManager) AddHandler(name string, handler slog.Handler) {
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

// RemoveHandler removes a [slog.Handler] from being managed.
func (m *slogManager) RemoveHandler(name string) {
	m.Lock()
	defer m.Unlock()

	delete(m.handlers, name)
}
