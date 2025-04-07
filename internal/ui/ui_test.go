package ui

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/desertwitch/gover/internal/queue"
	"github.com/desertwitch/gover/internal/schema"
)

type fakeShare struct {
	name string
}

func (s *fakeShare) GetName() string                          { return s.name }
func (s *fakeShare) GetUseCache() string                      { return "" }
func (s *fakeShare) GetCachePool() schema.Pool                { return nil }
func (s *fakeShare) GetCachePool2() schema.Pool               { return nil }
func (s *fakeShare) GetAllocator() string                     { return "" }
func (s *fakeShare) GetSplitLevel() int                       { return 0 }
func (s *fakeShare) GetSpaceFloor() uint64                    { return 0 }
func (s *fakeShare) GetDisableCOW() bool                      { return false }
func (s *fakeShare) GetIncludedDisks() map[string]schema.Disk { return nil }

// TestTeaUI is an integration test for the command-line user interface.
func TestTeaUI(t *testing.T) {
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

	go func() {
		// Simulate some progress work for the UI to render.
		for {
			time.Sleep(time.Millisecond)
			if handler.Initialized.Load() {
				time.Sleep(time.Millisecond)
				handler.queueManager.EvaluationManager.Enqueue(
					&schema.Moveable{
						Share: &fakeShare{name: "share1"},
					},
					&schema.Moveable{
						Share: &fakeShare{name: "share2"},
					},
					&schema.Moveable{
						Share: &fakeShare{name: "share3"},
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

	if err := handler.Launch(); err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("UI generated no output at all")
	}

	by := buf.Bytes()

	if !bytes.Contains(by, []byte("log1")) {
		t.Fatal("UI did not show the first log message sent (via program.Send)")
	}

	if !bytes.Contains(by, []byte("log2")) {
		t.Fatal("UI did not show the second log message sent (via LogWriter)")
	}

	if !bytes.Contains(by, []byte("Finished")) {
		t.Fatal("UI did not update the progress panels.")
	}
}

// TestTeaUI_Ctrl_C is an integration test for the command-line user interface.
// A Ctrl+C keypress is simulated, which should trigger upstream Context
// cancellation for signalling application teardown.
func TestTeaUI_Ctrl_C(t *testing.T) {
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

	if err == nil {
		t.Fatalf("Expected %v, got nil", context.Canceled)
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Expected %v, got %v", context.Canceled, err)
	}

	if buf.Len() == 0 {
		t.Fatal("UI generated no output at all")
	}
}
