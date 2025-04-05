package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// logMsg is a regular string containing a log message. It is typed for
// identification as [tea.Msg] within a [tea.Program].
type logMsg string

// TeaLogWriter is an implementation of an [io.Writer], for use inside a
// [slog.Handler], that sends any logs to a [tea.Program] as [tea.Msg].
type TeaLogWriter struct {
	program  *tea.Program
	doneChan chan struct{}
	logChan  chan logMsg
}

// NewTeaLogWriter returns a pointer to a new [TeaLogWriter]. It also starts the
// internal log processing function, which should eventually be stopped e.g.
// with a deferred [TeaLogWriter.Stop] call.
func NewTeaLogWriter(program *tea.Program) *TeaLogWriter {
	wr := &TeaLogWriter{
		program:  program,
		doneChan: make(chan struct{}),
		logChan:  make(chan logMsg, 1000), //nolint:mnd
	}

	go wr.processLogs()

	return wr
}

// Stop destroys the [TeaLogWriter] and stops any log message processing. This
// should be called when no more logs are actively being sent, as any in-flight
// or late logs will be discarded after calling this method.
func (wr *TeaLogWriter) Stop() {
	close(wr.doneChan)
}

// processLogs sends any received logs to the [tea.Program] as [tea.Msg]. The
// logs are received from the internal buffered channel, filled by
// [TeaLogWriter.Write].
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

// Write receives a byte slice containing a log message from e.g. a
// [slog.Handler]. It is interally sent into a buffered channel, received by
// [TeaLogWriter.processLogs].
func (wr *TeaLogWriter) Write(p []byte) (int, error) {
	logStr := string(p)

	select {
	case <-wr.doneChan:
	case wr.logChan <- logMsg(logStr):
	}

	return len(p), nil
}
