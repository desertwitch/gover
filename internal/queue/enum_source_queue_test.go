package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewEnumerationSourceQueue_Success tests the factory function.
func TestNewEnumerationSourceQueue_Success(t *testing.T) {
	t.Parallel()

	queue := NewEnumerationSourceQueue()

	require.NotNil(t, queue, "NewEnumerationSourceQueue() should return a non-nil value")
	require.NotNil(t, queue.GenericQueue, "NewEnumerationSourceQueue() should initialize the embedded GenericQueue")
}

// TestEnumerationTaskRun_Success tests running an EnumerationTask's function.
func TestEnumerationTaskRun_Success(t *testing.T) {
	t.Parallel()

	task := &EnumerationTask{
		Share:    nil,
		Source:   nil,
		Function: func() int { return 42 },
	}

	result := task.Run()
	assert.Equal(t, 42, result, "EnumerationTask.Run() should return the result of its Function")
}

// TestEnumerationSourceQueueProgress_Success tests the progress method.
func TestEnumerationSourceQueueProgress_Success(t *testing.T) {
	t.Parallel()

	queue := NewEnumerationSourceQueue()

	progress := queue.Progress()

	assert.Equal(t, "tasks/sec", progress.TransferSpeedUnit, "Progress should use 'tasks/sec' as the transfer speed unit")
}
