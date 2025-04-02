package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type teaLogWriter struct {
	program  *tea.Program
	doneChan chan struct{}
	logChan  chan logMsg
}

//nolint:mnd
func newTeaLogWriter(program *tea.Program) *teaLogWriter {
	wr := &teaLogWriter{
		program:  program,
		doneChan: make(chan struct{}),
		logChan:  make(chan logMsg, 1000),
	}

	go wr.processLogs()

	return wr
}

func (wr *teaLogWriter) Stop() {
	close(wr.doneChan)
}

func (wr *teaLogWriter) processLogs() {
	for {
		select {
		case <-wr.doneChan:
			return
		case msg := <-wr.logChan:
			wr.program.Send(msg)
		}
	}
}

func (wr *teaLogWriter) Write(p []byte) (int, error) {
	logStr := string(p)

	select {
	case <-wr.doneChan:
	case wr.logChan <- logMsg(logStr):
	}

	return len(p), nil
}
