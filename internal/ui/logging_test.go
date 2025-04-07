package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeProgram is a fake implementation of teaProgramProvider. It collects all
// messages sent via its Send method.
type fakeProgram struct {
	msgs chan tea.Msg
}

func newFakeProgram() *fakeProgram {
	return &fakeProgram{
		msgs: make(chan tea.Msg, 100),
	}
}

func (fp *fakeProgram) Send(msg tea.Msg) {
	fp.msgs <- msg
}

// TestTeaLogWriter_Write_Table verifies that calls to Write send the expected
// messages.
func TestTeaLogWriter_Write_Table(t *testing.T) {
	t.Parallel()

	fp := newFakeProgram()
	writer := NewTeaLogWriter(fp)
	defer writer.Stop()

	testCases := []struct {
		name  string
		input string
	}{
		{"Success_EmptyMessage", ""},
		{"Success_ShortMessage", "log"},
		{"Success_LongMessage", "this is a longer log message"},
		{"Success_UnicodeMessage", "this is a Japanese message - 日本!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, err := writer.Write([]byte(tc.input))
			require.NoError(t, err)
			require.Equal(t, len(tc.input), n)

			select {
			case got := <-fp.msgs:
				assert.Equal(t, LogMsg(tc.input), got)
			case <-time.After(300 * time.Millisecond):
				t.Fatalf("timeout waiting for log message in case: %s", tc.name)
			}
		})
	}
}

// TestTeaLogWriter_Stop verifies that after Stop is called, subsequent Write
// calls do not send messages.
func TestTeaLogWriter_Stop(t *testing.T) {
	t.Parallel()

	fp := newFakeProgram()
	writer := NewTeaLogWriter(fp)

	_, _ = writer.Write([]byte("first message"))

	time.Sleep(50 * time.Millisecond)
	writer.Stop()
	time.Sleep(50 * time.Millisecond)

	_, _ = writer.Write([]byte("second message"))
	_, _ = writer.Write([]byte("third message"))
	_, _ = writer.Write([]byte("fourth message"))

	var msgs []string
drainLoop:
	for {
		select {
		case m := <-fp.msgs:
			if lm, ok := m.(LogMsg); ok {
				msgs = append(msgs, string(lm))
			}
		case <-time.After(300 * time.Millisecond):
			break drainLoop
		}
	}

	assert.Contains(t, msgs, "first message", "expected first message to be delivered")
	for _, s := range msgs {
		assert.NotEqual(t, "second message", s, "expected second message not to be delivered")
	}
}
