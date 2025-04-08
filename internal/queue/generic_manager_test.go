package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FakeQueue is a generic queue for testing.
type FakeQueue struct {
	*GenericQueue[int]
}

// NewFakeQueue returns a pointer to a new [FakeQueue].
func NewFakeQueue() *FakeQueue {
	return &FakeQueue{
		GenericQueue: NewGenericQueue[int](),
	}
}

var keyFunc = func(i int) string {
	if i <= 0 {
		return "A"
	}
	if i > 0 && i <= 5 {
		return "B"
	}
	if i > 5 && i <= 10 {
		return "C"
	}

	return "D"
}

// TestNewGenericManager_Success tests the factory function.
func TestNewGenericManager_Success(t *testing.T) {
	t.Parallel()

	qm := NewGenericManager[int, *FakeQueue]()
	require.NotNil(t, qm.queues, "queue map should not be nil")
}

// TestNewGenericManager_GetSuccessful tests getting successful items.
func TestNewGenericManager_GetSuccessful(t *testing.T) {
	t.Parallel()

	qm := NewGenericManager[int, *FakeQueue]()
	require.NotNil(t, qm.queues, "queue map should not be nil")

	qm.Enqueue(0, keyFunc, NewFakeQueue)
	qm.Enqueue(4, keyFunc, NewFakeQueue)
	qm.Enqueue(8, keyFunc, NewFakeQueue)
	qm.Enqueue(20, keyFunc, NewFakeQueue)
	require.Len(t, qm.queues, 4, "should have created 4 queues")

	item, ok := qm.queues["A"].Dequeue()
	assert.True(t, ok)
	assert.Equal(t, 0, item)
	qm.queues["A"].SetSuccess(item)

	item, ok = qm.queues["B"].Dequeue()
	assert.True(t, ok)
	assert.Equal(t, 4, item)
	qm.queues["B"].SetSuccess(item)

	item, ok = qm.queues["C"].Dequeue()
	assert.True(t, ok)
	assert.Equal(t, 8, item)
	qm.queues["C"].SetProcessing(item)

	successful := qm.GetSuccessful()
	require.Len(t, successful, 2, "var should have 2 successful elements")

	qm.queues["C"].SetSuccess(8)
	require.Len(t, qm.GetSuccessful(), 3, "qm should have 3 successful elements")

	assert.NotEqual(t, successful, qm.GetSuccessful(), "returned slice should be a copy")
}

// TestGenericManager_Enqueue_Success tests enqueuing.
func TestGenericManager_Enqueue_Success(t *testing.T) {
	t.Parallel()

	qm := NewGenericManager[int, *FakeQueue]()
	require.NotNil(t, qm.queues, "queue map should not be nil")

	qm.Enqueue(0, keyFunc, NewFakeQueue)
	qm.Enqueue(4, keyFunc, NewFakeQueue)
	qm.Enqueue(4, keyFunc, NewFakeQueue)
	qm.Enqueue(8, keyFunc, NewFakeQueue)
	qm.Enqueue(8, keyFunc, NewFakeQueue)

	assert.Len(t, qm.queues, 3, "should have created 3 queues")
}

// TestGenericManager_GetQueues_Success tests getting queues.
func TestGenericManager_GetQueues_Success(t *testing.T) {
	t.Parallel()

	qm := NewGenericManager[int, *FakeQueue]()
	require.NotNil(t, qm.queues, "queue map should not be nil")

	qm.Enqueue(0, keyFunc, NewFakeQueue)
	qm.Enqueue(4, keyFunc, NewFakeQueue)
	qm.Enqueue(4, keyFunc, NewFakeQueue)
	qm.Enqueue(8, keyFunc, NewFakeQueue)
	qm.Enqueue(8, keyFunc, NewFakeQueue)

	queues := qm.GetQueues()

	qm.Enqueue(20, keyFunc, NewFakeQueue)

	assert.NotEqual(t, qm.queues, queues)
	assert.Len(t, qm.queues, 4, "should have created 4 queues")
	assert.Len(t, queues, 3, "should have returned 3 queues (copy)")
}

// TestGenericManager_GetQueues_Success_Nil tests getting queues when nil.
func TestGenericManager_GetQueues_Success_Nil(t *testing.T) {
	t.Parallel()

	qm := &GenericManager[int, FakeQueue]{}
	require.Nil(t, qm.GetQueues(), "returned queues should be nil")
}

// TestGenericManager_Progress_Success tests getting progress.
func TestGenericManager_Progress_Success(t *testing.T) {
	t.Parallel()

	qm := NewGenericManager[int, *FakeQueue]()
	require.NotNil(t, qm.queues, "queue map should not be nil")

	progress := qm.Progress()
	assert.False(t, progress.HasStarted)
	assert.False(t, progress.HasFinished)
	assert.InDelta(t, 0.0, progress.ProgressPct, 0)
	assert.Equal(t, 0, progress.TotalItems)
	assert.Equal(t, 0, progress.ProcessedItems)
	assert.Equal(t, 0, progress.SuccessItems)
	assert.Equal(t, 0, progress.SkippedItems)
	assert.Equal(t, 0, progress.InProgressItems)

	qm.Enqueue(0, keyFunc, NewFakeQueue)
	qm.Enqueue(4, keyFunc, NewFakeQueue)
	qm.Enqueue(8, keyFunc, NewFakeQueue)
	qm.Enqueue(20, keyFunc, NewFakeQueue)
	require.Len(t, qm.queues, 4, "should have created 4 queues")

	qm.queues["A"].Dequeue()
	qm.queues["A"].SetSuccess(0)
	qm.queues["B"].Dequeue()
	qm.queues["B"].SetSkipped(4)
	qm.queues["C"].Dequeue()
	qm.queues["C"].SetProcessing(8)

	progress = qm.Progress()
	assert.True(t, progress.HasStarted)
	assert.False(t, progress.HasFinished)
	assert.InDelta(t, 50.0, progress.ProgressPct, 0)
	assert.Equal(t, 4, progress.TotalItems)
	assert.Equal(t, 2, progress.ProcessedItems)
	assert.Equal(t, 1, progress.SuccessItems)
	assert.Equal(t, 1, progress.SkippedItems)
	assert.Equal(t, 1, progress.InProgressItems)
	assert.NotZero(t, progress.StartTime, "start time should not be zero")
	assert.NotZero(t, progress.ETA, "eta should not be zero")

	qm.queues["C"].SetSuccess(8)
	qm.queues["D"].Dequeue()
	qm.queues["D"].SetSuccess(20)

	progress = qm.Progress()
	assert.True(t, progress.HasStarted)
	assert.True(t, progress.HasFinished)
	assert.NotZero(t, progress.FinishTime, "finish time should not be zero")
	assert.InDelta(t, 100.0, progress.ProgressPct, 0)
}
