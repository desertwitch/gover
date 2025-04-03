package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type TeaLogWriter struct {
	program  *tea.Program
	doneChan chan struct{}
	logChan  chan logMsg
}

//nolint:mnd
func NewTeaLogWriter(program *tea.Program) *TeaLogWriter {
	wr := &TeaLogWriter{
		program:  program,
		doneChan: make(chan struct{}),
		logChan:  make(chan logMsg, 1000),
	}

	go wr.processLogs()

	return wr
}

func (wr *TeaLogWriter) Stop() {
	close(wr.doneChan)
}

func (wr *TeaLogWriter) processLogs() {
	for {
		select {
		case <-wr.doneChan:
			return
		case msg := <-wr.logChan:
			wr.program.Send(msg)
		}
	}
}

func (wr *TeaLogWriter) Write(p []byte) (int, error) {
	logStr := string(p)

	select {
	case <-wr.doneChan:
	case wr.logChan <- logMsg(logStr):
	}

	return len(p), nil
}
