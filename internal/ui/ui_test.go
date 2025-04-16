package ui

import (
	"bytes"
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/desertwitch/gover/internal/queue"
	"github.com/desertwitch/gover/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewHandler_Success tests the factory function.
func TestNewHandler_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	cancel := func() {}

	mockQueueManager := &queue.Manager{}

	handler := NewHandler(ctx, cancel, mockQueueManager)
	defer handler.LogWriter.Stop()

	require.NotNil(t, handler, "Handler should not be nil")
	assert.Equal(t, mockQueueManager, handler.queueManager, "QueueManager should be correctly assigned")
	require.NotNil(t, handler.program, "Program should be initialized")
	require.NotNil(t, handler.LogWriter, "LogWriter should be initialized")
}

// TestTeaUI_Integration is an integration test for the command-line user
// interface.
func TestTeaUI_Integration(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	var in bytes.Buffer

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	handler := &Handler{queueManager: queue.NewManager()}
	model := NewTeaModel(handler, cancel)
	program := tea.NewProgram(model, tea.WithInput(&in), tea.WithOutput(&buf), tea.WithAltScreen(), tea.WithContext(ctx))

	handler.program = program
	handler.LogWriter = NewTeaLogWriter(handler.program)

	share1 := schema.NewMockShare(t)
	share2 := schema.NewMockShare(t)
	share3 := schema.NewMockShare(t)

	go func() {
		// Simulate some progress work for the UI to render.
		for {
			time.Sleep(time.Millisecond)
			if handler.Initialized.Load() {
				time.Sleep(time.Millisecond)
				handler.queueManager.EvaluationManager.Enqueue(
					&schema.Moveable{
						Share: share1,
					},
					&schema.Moveable{
						Share: share2,
					},
					&schema.Moveable{
						Share: share3,
					},
				)
				for _, q := range handler.queueManager.EvaluationManager.GetQueues() {
					_ = q.DequeueAndProcess(ctx, func(m *schema.Moveable) int {
						time.Sleep(100 * time.Millisecond)

						return queue.DecisionSuccess
					})
				}

				return
			}
		}
	}()

	go func() {
		// Simulate some fast-paced logs and key presses for the UI.
		for {
			time.Sleep(time.Millisecond)
			if handler.Initialized.Load() {
				program.Send(tea.WindowSizeMsg{Width: 200, Height: 200})
				time.Sleep(time.Millisecond)

				program.Send(LogMsg("log1"))
				time.Sleep(time.Millisecond)

				_, _ = handler.LogWriter.Write([]byte("log2"))
				time.Sleep(time.Millisecond)

				for range 150 {
					_, _ = handler.LogWriter.Write([]byte("fast logs"))
				}
				time.Sleep(time.Millisecond)

				program.Send(tea.WindowSizeMsg{Width: 200, Height: 250})

				time.Sleep(3 * time.Second)
				program.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

				return
			}
			if handler.Failed.Load() {
				return
			}
		}
	}()

	err := handler.Launch()

	require.NoError(t, err, "Handler launch should succeed")
	require.NotZero(t, buf.Len(), "UI should generate output")

	by := buf.Bytes()

	assert.Contains(t, string(by), "log1", "UI should show the first log message")
	assert.Contains(t, string(by), "log2", "UI should show the second log message")
	assert.Contains(t, string(by), "Finished", "UI should show progress as finished")

	share1.AssertExpectations(t)
	share2.AssertExpectations(t)
	share3.AssertExpectations(t)
}

// TestTeaUI_Integration_Ctrl_C is an integration test for the command-line user
// interface. A Ctrl+C keypress is simulated, which should trigger upstream
// Context cancellation for signalling application teardown.
func TestTeaUI_Integration_Ctrl_C(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	var in bytes.Buffer

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	handler := &Handler{queueManager: queue.NewManager()}

	model := NewTeaModel(handler, cancel)
	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithInput(&in), tea.WithOutput(&buf), tea.WithContext(ctx))

	handler.program = program
	handler.LogWriter = NewTeaLogWriter(handler.program)

	go func() {
		for {
			time.Sleep(time.Millisecond)
			if handler.Initialized.Load() {
				program.Send(tea.KeyMsg{Type: tea.KeyCtrlC})

				return
			}
			if handler.Failed.Load() {
				return
			}
		}
	}()

	err := handler.Launch()

	require.ErrorIs(t, err, context.Canceled, "UI should cancel external context")
	require.NotZero(t, buf.Len(), "UI should generate output")
}
