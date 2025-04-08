package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewEnumerationSourceQueue_Success tests the factory function.
func TestNewEvaluationShareQueue_Success(t *testing.T) {
	t.Parallel()

	queue := NewEvaluationShareQueue()

	require.NotNil(t, queue, "NewEvaluationShareQueue() should return a non-nil value")
	require.NotNil(t, queue.GenericQueue, "NewEvaluationShareQueue() should initialize the embedded GenericQueue")
}

// TestEvaluationShareQueueProgress_Success tests the progress method.
func TestEvaluationShareQueueProgress_Success(t *testing.T) {
	t.Parallel()

	queue := NewEvaluationShareQueue()

	progress := queue.Progress()

	assert.Equal(t, "items/sec", progress.TransferSpeedUnit, "Progress should use 'items/sec' as the transfer speed unit")
}
