package ui

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type teaLogWriter struct {
	sync.RWMutex
	program  *tea.Program
	sendChan chan logMsg
}

//nolint:mnd
func newTeaLogWriter() *teaLogWriter {
	return &teaLogWriter{
		sendChan: make(chan logMsg, 1000),
	}
}

func (wr *teaLogWriter) processLogs() {
	wr.RLock()
	program := wr.program
	sendChan := wr.sendChan
	wr.RUnlock()

	for msg := range sendChan {
		program.Send(msg)
	}
}

func (wr *teaLogWriter) Init() {
	go wr.processLogs()
}

func (wr *teaLogWriter) Stop() {
	wr.Lock()
	defer wr.Unlock()

	sendChan := wr.sendChan
	wr.sendChan = nil
	close(sendChan)
}

func (wr *teaLogWriter) SetProgram(p *tea.Program) {
	wr.Lock()
	defer wr.Unlock()

	wr.program = p
}

func (wr *teaLogWriter) Write(p []byte) (int, error) {
	logStr := string(p)

	wr.RLock()
	program := wr.program
	sendChan := wr.sendChan
	wr.RUnlock()

	if program != nil && sendChan != nil {
		select {
		case sendChan <- logMsg(logStr):
		case <-time.After(time.Second):
		}
	}

	return len(p), nil
}
