package ui

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type teaLogWriter struct {
	program  *tea.Program
	sendChan chan logMsg
	mu       sync.RWMutex
}

//nolint:mnd
func newTeaLogWriter() *teaLogWriter {
	return &teaLogWriter{
		sendChan: make(chan logMsg, 1000),
	}
}

func (h *teaLogWriter) processLogs() {
	h.mu.RLock()
	program := h.program
	sendChan := h.sendChan
	h.mu.RUnlock()

	for msg := range sendChan {
		program.Send(msg)
	}
}

func (h *teaLogWriter) Start() {
	go h.processLogs()
}

func (h *teaLogWriter) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	sendChan := h.sendChan
	h.sendChan = nil
	close(sendChan)
}

func (h *teaLogWriter) SetProgram(p *tea.Program) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.program = p
}

func (h *teaLogWriter) Write(p []byte) (int, error) {
	logStr := string(p)

	h.mu.RLock()
	program := h.program
	sendChan := h.sendChan
	h.mu.RUnlock()

	if program != nil && sendChan != nil {
		select {
		case sendChan <- logMsg(logStr):
		case <-time.After(time.Second):
		}
	}

	return len(p), nil
}
