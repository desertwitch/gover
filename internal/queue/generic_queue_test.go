package queue

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewGenericQueue_Success tests the queue factory function.
func TestNewGenericQueue_Success(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[string]()

	assert.NotNil(t, q)
	assert.Empty(t, q.items)
	assert.Empty(t, q.success)
	assert.Empty(t, q.skipped)
	assert.NotNil(t, q.inProgress)
	assert.Equal(t, 0, q.head)
	assert.False(t, q.hasStarted)
	assert.False(t, q.hasFinished)
}

// TestEnqueueDequeue_Success tests enqueueing and dequeueing.
func TestEnqueueDequeue_Success(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[string]()

	q.Enqueue("item1", "item2", "item3")

	assert.Len(t, q.items, 3)
	assert.Equal(t, []string{"item1", "item2", "item3"}, q.items)

	item, ok := q.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, "item1", item)
	assert.Equal(t, 1, q.head)

	item, ok = q.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, "item2", item)
	assert.Equal(t, 2, q.head)

	item, ok = q.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, "item3", item)
	assert.Equal(t, 3, q.head)

	item, ok = q.Dequeue()
	assert.False(t, ok)
	assert.Equal(t, "", item)
}

// TestHasRemainingItems_Success tests for remaining items.
func TestHasRemainingItems_Success(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[int]()

	assert.False(t, q.HasRemainingItems())

	q.Enqueue(1, 2, 3)

	assert.True(t, q.HasRemainingItems())

	q.Dequeue()
	q.Dequeue()
	q.Dequeue()

	assert.False(t, q.HasRemainingItems())
}

// TestGetSuccessful_Success tests returning the success items.
func TestGetSuccessful_Success(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[int]()

	success := q.GetSuccessful()
	assert.Empty(t, success)

	q.SetSuccess(1, 2, 3)
	success = q.GetSuccessful()
	assert.Equal(t, []int{1, 2, 3}, success)

	// Verify we get a copy, not a reference
	success[0] = 999
	newSuccess := q.GetSuccessful()
	assert.Equal(t, []int{1, 2, 3}, newSuccess)
}

// TestSetProcessing_Success tests setting items as processing.
func TestSetProcessing_Success(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[string]()

	assert.Empty(t, q.inProgress)

	q.SetProcessing("item1", "item2")
	assert.Len(t, q.inProgress, 2)

	_, exists := q.inProgress["item1"]
	assert.True(t, exists)

	_, exists = q.inProgress["item2"]
	assert.True(t, exists)
}

// TestSetSuccess_Success tests setting items as successful.
func TestSetSuccess_Success(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[string]()

	q.SetProcessing("item1", "item2")
	assert.Len(t, q.inProgress, 2)

	q.SetSuccess("item1")
	assert.Contains(t, q.success, "item1")

	assert.Len(t, q.inProgress, 1)

	_, exists := q.inProgress["item1"]
	assert.False(t, exists)

	_, exists = q.inProgress["item2"]
	assert.True(t, exists)
}

// TestSetSkipped_Success tests setting items as skipped.
func TestSetSkipped_Success(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[string]()

	q.SetProcessing("item1", "item2")
	assert.Len(t, q.inProgress, 2)

	q.SetSkipped("item1")
	assert.Contains(t, q.skipped, "item1")

	assert.Len(t, q.inProgress, 1)

	_, exists := q.inProgress["item1"]
	assert.False(t, exists)

	_, exists = q.inProgress["item2"]
	assert.True(t, exists)
}

// TestProgress_Success tests the progress being reported.
func TestProgress_Success(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[int]()

	progress := q.Progress()
	assert.False(t, progress.HasStarted)
	assert.False(t, progress.HasFinished)
	assert.Zero(t, progress.StartTime, "start time should be zero")
	assert.Zero(t, progress.FinishTime, "finish time should be zero")
	assert.Zero(t, progress.ETA, "eta should be zero")
	assert.Zero(t, progress.TimeLeft, "time left should be zero")
	assert.InDelta(t, 0.0, progress.ProgressPct, 0)
	assert.Equal(t, 0, progress.TotalItems)
	assert.Equal(t, 0, progress.ProcessedItems)
	assert.Equal(t, 0, progress.SuccessItems)
	assert.Equal(t, 0, progress.SkippedItems)
	assert.Equal(t, 0, progress.InProgressItems)

	q.Enqueue(1, 2, 3, 4)
	q.Dequeue()
	q.SetSuccess(1)
	q.Dequeue()
	q.SetSkipped(2)
	q.Dequeue()
	q.SetSkipped(3)

	progress = q.Progress()
	assert.True(t, progress.HasStarted)
	assert.False(t, progress.HasFinished)
	assert.NotZero(t, progress.StartTime, "start time should not be zero")
	assert.Zero(t, progress.FinishTime, "finish time should be zero")
	assert.NotZero(t, progress.ETA, "eta should not be zero")
	assert.NotZero(t, progress.TimeLeft, "time left should not be zero")
	assert.InDelta(t, 75.0, progress.ProgressPct, 0)
	assert.Equal(t, 4, progress.TotalItems)
	assert.Equal(t, 3, progress.ProcessedItems)
	assert.Equal(t, 1, progress.SuccessItems)
	assert.Equal(t, 2, progress.SkippedItems)
	assert.Equal(t, 0, progress.InProgressItems)

	q.Dequeue()
	q.SetSuccess(4)

	progress = q.Progress()
	assert.True(t, progress.HasStarted)
	assert.True(t, progress.HasFinished)
	assert.NotZero(t, progress.StartTime, "start time should not be zero")
	assert.NotZero(t, progress.FinishTime, "finish time should not be zero")
	assert.Zero(t, progress.ETA, "eta should be zero")
	assert.Zero(t, progress.TimeLeft, "time left should be zero")
	assert.Equal(t, 4, progress.TotalItems)
	assert.Equal(t, 4, progress.ProcessedItems)
	assert.Equal(t, 2, progress.SuccessItems)
	assert.Equal(t, 2, progress.SkippedItems)
	assert.Equal(t, 0, progress.InProgressItems)
	assert.InDelta(t, 100.0, progress.ProgressPct, 0)
}

// TestDequeueAndProcess_Success tests sequential processing.
func TestDequeueAndProcess_Success(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[string]()
	q.Enqueue("success", "skip", "requeue", "success2")

	attempts := make(map[string]int)
	processFunc := func(item string) int {
		attempts[item]++

		switch item {
		case "success", "success2":
			return DecisionSuccess
		case "skip":
			return DecisionSkipped
		case "requeue":
			if attempts[item] < 2 {
				return DecisionRequeue
			}

			return DecisionSuccess
		default:
			return DecisionSkipped
		}
	}

	ctx := t.Context()
	err := q.DequeueAndProcess(ctx, processFunc)

	require.NoError(t, err)

	assert.False(t, q.HasRemainingItems())
	assert.Len(t, q.success, 3)
	assert.Len(t, q.skipped, 1)
	assert.Equal(t, 2, attempts["requeue"])

	assert.False(t, q.HasRemainingItems())
}

// TestDequeueAndProcess_Fail_CtxCancel tests in-flight cancellation during
// sequential processing.
func TestDequeueAndProcess_Fail_CtxCancel(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[int]()
	q.Enqueue(1, 2, 3, 4, 5)

	ctx, cancel := context.WithCancel(t.Context())

	processFunc := func(item int) int {
		if item == 3 {
			cancel()
		}

		return DecisionSuccess
	}

	err := q.DequeueAndProcess(ctx, processFunc)

	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)

	assert.True(t, q.HasRemainingItems())
}

// TestDequeueAndProcessConc_Success tests concurrent processing.
func TestDequeueAndProcessConc_Success(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[int]()

	const itemCount = 100
	for i := 1; i <= itemCount; i++ {
		q.Enqueue(i)
	}

	var successCount atomic.Int32
	var skippedCount atomic.Int32
	var requeueCount atomic.Int32
	var processedCount atomic.Int32
	var inFlightCount atomic.Int32

	processFunc := func(item int) int {
		processedCount.Add(1)

		inFlightCount.Add(1)
		defer inFlightCount.Add(-1)

		// Artificial delay to test concurrency
		time.Sleep(time.Millisecond * 5)

		if item%5 == 0 {
			skippedCount.Add(1)

			return DecisionSkipped
		}

		if item%10 == 0 {
			if requeueCount.Load() < 5 {
				requeueCount.Add(1)

				return DecisionRequeue
			}
		}

		require.LessOrEqual(t, int(inFlightCount.Load()), 10, "inFlight should not exceed maxWorkers")

		successCount.Add(1)

		return DecisionSuccess
	}

	ctx := t.Context()
	err := q.DequeueAndProcessConc(ctx, 10, processFunc)

	require.NoError(t, err)

	assert.Len(t, q.success, int(successCount.Load()))
	assert.Len(t, q.skipped, int(skippedCount.Load()))
	assert.Empty(t, q.inProgress)
	assert.Equal(t, (len(q.success) + len(q.skipped)), int(successCount.Load()+skippedCount.Load()))

	assert.False(t, q.HasRemainingItems())
}

// TestDequeueAndProcessConc_Fail_CtxCancel tests mid-flight context
// cancellation during concurrent processing.
func TestDequeueAndProcessConc_Fail_CtxCancel(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[int]()

	const itemCount = 50
	for i := 1; i <= itemCount; i++ {
		q.Enqueue(i)
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond*50)
	defer cancel()

	processFunc := func(item int) int {
		time.Sleep(time.Millisecond * 10)

		return DecisionSuccess
	}

	err := q.DequeueAndProcessConc(ctx, 5, processFunc)

	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

// TestRequeueAndReprocess_Success tests requeuing of items.
func TestRequeueAndReprocess_Success(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[string]()
	q.Enqueue("item1", "item2", "item3", "requeueMe")

	var mu sync.Mutex
	attempts := make(map[string]int)

	processFunc := func(item string) int {
		mu.Lock()
		defer mu.Unlock()

		attempts[item]++

		if item == "requeueMe" {
			if attempts[item] < 3 {
				return DecisionRequeue
			}
		}

		return DecisionSuccess
	}

	ctx := t.Context()
	err := q.DequeueAndProcessConc(ctx, 2, processFunc)

	require.NoError(t, err)

	assert.Len(t, q.success, 4)
	assert.Equal(t, 3, attempts["requeueMe"])
	assert.Equal(t, 1, attempts["item1"])
	assert.Equal(t, 1, attempts["item2"])
	assert.Equal(t, 1, attempts["item3"])

	assert.False(t, q.HasRemainingItems())
}

// TestEnqueueAfterFinish_Success tests enqueueing after queue finish.
func TestEnqueueAfterFinish_Success(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[int]()

	q.Dequeue()
	assert.False(t, q.hasFinished)

	q.Enqueue(1, 2, 3)

	for q.HasRemainingItems() {
		item, ok := q.Dequeue()
		assert.True(t, ok)
		q.SetSuccess(item)
	}

	assert.True(t, q.hasStarted)
	assert.True(t, q.hasFinished)

	assert.Len(t, q.success, 3)

	q.Enqueue(1, 2, 3)
	q.Dequeue()
	assert.False(t, q.hasFinished)
}

// TestQueueWithCustomType_Success tests the queue with a complex type.
func TestQueueWithCustomType_Success(t *testing.T) {
	t.Parallel()

	type CustomItem struct {
		ID   int
		Name string
	}

	q := NewGenericQueue[CustomItem]()
	items := []CustomItem{
		{ID: 1, Name: "Item 1"},
		{ID: 2, Name: "Item 2"},
		{ID: 3, Name: "Item 3"},
	}

	q.Enqueue(items...)
	assert.Len(t, q.items, 3)

	item, ok := q.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, 1, item.ID)
	assert.Equal(t, "Item 1", item.Name)

	assert.False(t, q.hasFinished)
}

// TestProgressCalculation_Success tests progress calculations.
func TestProgressCalculation_Success(t *testing.T) {
	t.Parallel()

	q := NewGenericQueue[int]()

	const itemCount = 100
	for i := 1; i <= itemCount; i++ {
		q.Enqueue(i)
	}

	for range 50 {
		item, ok := q.Dequeue()
		require.True(t, ok)
		q.SetSuccess(item)
	}

	assert.True(t, q.hasStarted)

	progress := q.Progress()
	assert.InDelta(t, 50.0, progress.ProgressPct, 0)
	assert.True(t, progress.HasStarted)
	assert.False(t, progress.HasFinished)
	assert.Equal(t, itemCount, progress.TotalItems)
	assert.Equal(t, 50, progress.ProcessedItems)
	assert.Equal(t, 50, progress.SuccessItems)

	// The following are time-dependent and may be flaky in CI environments So
	// we only check that they're set, not their precise values
	assert.NotZero(t, progress.TransferSpeed)
	assert.Equal(t, "items/sec", progress.TransferSpeedUnit)
	assert.NotZero(t, progress.TimeLeft)
	assert.NotZero(t, progress.ETA)

	for q.HasRemainingItems() {
		item, ok := q.Dequeue()
		require.True(t, ok)
		q.SetSuccess(item)
	}

	progress = q.Progress()
	assert.InDelta(t, 100.0, progress.ProgressPct, 0)
	assert.True(t, progress.HasStarted)
	assert.True(t, progress.HasFinished)
	assert.Equal(t, itemCount, progress.TotalItems)
	assert.Equal(t, itemCount, progress.ProcessedItems)
	assert.Equal(t, itemCount, progress.SuccessItems)
	assert.Zero(t, progress.TimeLeft)
	assert.Zero(t, progress.ETA)
}
