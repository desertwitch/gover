package queue

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenericQueue_Success(t *testing.T) {
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

func TestEnqueueDequeue_Success(t *testing.T) {
	q := NewGenericQueue[string]()

	q.Enqueue("item1", "item2", "item3")

	assert.Equal(t, 3, len(q.items))
	assert.Equal(t, []string{"item1", "item2", "item3"}, q.items)

	item, ok := q.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, "item1", item)
	assert.Equal(t, 1, q.head)

	assert.True(t, q.hasStarted)
	assert.False(t, q.hasFinished)

	item, ok = q.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, "item2", item)
	assert.Equal(t, 2, q.head)

	item, ok = q.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, "item3", item)
	assert.Equal(t, 3, q.head)
	assert.True(t, q.hasFinished)

	item, ok = q.Dequeue()
	assert.False(t, ok)
	assert.Equal(t, "", item)
	assert.True(t, q.hasFinished)
}

func TestHasRemainingItems_Success(t *testing.T) {
	q := NewGenericQueue[int]()

	assert.False(t, q.HasRemainingItems())

	q.Enqueue(1, 2, 3)

	assert.True(t, q.HasRemainingItems())

	q.Dequeue()
	assert.True(t, q.hasStarted)

	q.Dequeue()
	q.Dequeue()
	assert.False(t, q.HasRemainingItems())
}

func TestGetSuccessful_Success(t *testing.T) {
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

func TestSetProcessing_Success(t *testing.T) {
	q := NewGenericQueue[string]()

	assert.Empty(t, q.inProgress)

	q.SetProcessing("item1", "item2")
	assert.Len(t, q.inProgress, 2)
	_, exists := q.inProgress["item1"]
	assert.True(t, exists)
	_, exists = q.inProgress["item2"]
	assert.True(t, exists)
}

func TestSetSuccess_Success(t *testing.T) {
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

func TestSetSkipped_Success(t *testing.T) {
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

func TestProgress_Success(t *testing.T) {
	q := NewGenericQueue[int]()

	progress := q.Progress()
	assert.False(t, progress.HasStarted)
	assert.False(t, progress.HasFinished)
	assert.Equal(t, 0.0, progress.ProgressPct)
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
	q.SetProcessing(3)

	progress = q.Progress()
	assert.True(t, progress.HasStarted)
	assert.False(t, progress.HasFinished)
	assert.Equal(t, 75.0, progress.ProgressPct)
	assert.Equal(t, 4, progress.TotalItems)
	assert.Equal(t, 3, progress.ProcessedItems)
	assert.Equal(t, 1, progress.SuccessItems)
	assert.Equal(t, 1, progress.SkippedItems)
	assert.Equal(t, 1, progress.InProgressItems)

	q.Dequeue()
	q.SetSuccess(4)

	progress = q.Progress()
	assert.True(t, progress.HasStarted)
	assert.True(t, progress.HasFinished)
	assert.Equal(t, 100.0, progress.ProgressPct)
}

func TestDequeueAndProcess_Success(t *testing.T) {
	q := NewGenericQueue[string]()
	q.Enqueue("success", "skip", "requeue", "success2")

	processed := make(map[string]int)
	processFunc := func(item string) int {
		processed[item]++

		switch item {
		case "success", "success2":
			return DecisionSuccess
		case "skip":
			return DecisionSkipped
		case "requeue":
			if processed[item] < 2 {
				return DecisionRequeue
			}

			return DecisionSuccess
		default:
			return DecisionSkipped
		}
	}

	ctx := context.Background()
	err := q.DequeueAndProcess(ctx, processFunc)

	assert.NoError(t, err)
	assert.False(t, q.HasRemainingItems())
	assert.Equal(t, 3, len(q.success))
	assert.Equal(t, 1, len(q.skipped))
	assert.Equal(t, 2, processed["requeue"]) // Should be processed twice due to requeue

	assert.True(t, q.hasStarted)
	assert.True(t, q.hasFinished)

	assert.False(t, q.HasRemainingItems())
}

func TestDequeueAndProcess_Fail_CtxCancel(t *testing.T) {
	q := NewGenericQueue[int]()
	q.Enqueue(1, 2, 3, 4, 5)

	ctx, cancel := context.WithCancel(context.Background())

	processFunc := func(item int) int {
		if item == 3 {
			cancel()
		}

		return DecisionSuccess
	}

	err := q.DequeueAndProcess(ctx, processFunc)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))

	assert.True(t, q.HasRemainingItems())

	assert.True(t, q.hasStarted)
	assert.False(t, q.hasFinished)
}

func TestDequeueAndProcessConc_Success(t *testing.T) {
	q := NewGenericQueue[int]()

	const itemCount = 100
	for i := 1; i <= itemCount; i++ {
		q.Enqueue(i)
	}

	var successCount atomic.Int32
	var skippedCount atomic.Int32
	var requeueCount atomic.Int32
	var processedCount atomic.Int32

	processFunc := func(item int) int {
		processedCount.Add(1)

		// Artificial delay to test concurrency
		time.Sleep(time.Millisecond * 5)

		if item%5 == 0 {
			skippedCount.Add(1)

			return DecisionSkipped
		}

		if item%10 == 0 {
			if requeueCount.Load() < 5 { // Limit requeues to avoid test running too long
				requeueCount.Add(1)

				return DecisionRequeue
			}
		}

		successCount.Add(1)

		return DecisionSuccess
	}

	ctx := context.Background()
	err := q.DequeueAndProcessConc(ctx, 10, processFunc)

	assert.NoError(t, err)
	assert.False(t, q.HasRemainingItems())

	progress := q.Progress()
	assert.Equal(t, int(successCount.Load()), progress.SuccessItems)
	assert.Equal(t, int(skippedCount.Load()), progress.SkippedItems)
	assert.Equal(t, 0, progress.InProgressItems)

	totalProcessed := int(successCount.Load() + skippedCount.Load())
	assert.Equal(t, itemCount, totalProcessed)

	assert.True(t, q.hasFinished)
	assert.True(t, q.hasStarted)

	assert.False(t, q.HasRemainingItems())
}

func TestDequeueAndProcessConc_Fail_CtxCancel(t *testing.T) {
	q := NewGenericQueue[int]()

	const itemCount = 50
	for i := 1; i <= itemCount; i++ {
		q.Enqueue(i)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	processFunc := func(item int) int {
		time.Sleep(time.Millisecond * 10)

		return DecisionSuccess
	}

	err := q.DequeueAndProcessConc(ctx, 5, processFunc)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))

	assert.True(t, q.hasStarted)
	assert.False(t, q.hasFinished)
}

func TestRequeueAndReprocess_Success(t *testing.T) {
	q := NewGenericQueue[string]()
	q.Enqueue("item1", "item2", "item3", "requeueMe")

	attempts := make(map[string]int)
	processFunc := func(item string) int {
		attempts[item]++

		if item == "requeueMe" {
			if attempts[item] < 3 {
				return DecisionRequeue
			}
		}

		return DecisionSuccess
	}

	ctx := context.Background()
	err := q.DequeueAndProcessConc(ctx, 2, processFunc)

	assert.NoError(t, err)

	assert.Equal(t, 4, len(q.success))
	assert.Equal(t, 3, attempts["requeueMe"])
	assert.Equal(t, 1, attempts["item1"])
	assert.Equal(t, 1, attempts["item2"])
	assert.Equal(t, 1, attempts["item3"])

	assert.True(t, q.hasFinished)
	assert.True(t, q.hasStarted)

	assert.False(t, q.HasRemainingItems())
}

func TestEnqueueAfterFinish_Success(t *testing.T) {
	q := NewGenericQueue[int]()

	q.Dequeue()
	assert.False(t, q.hasFinished)

	q.Enqueue(1, 2, 3)

	for q.HasRemainingItems() {
		item, ok := q.Dequeue()
		assert.True(t, ok)
		q.SetSuccess(item)
	}

	assert.True(t, q.hasFinished)
	assert.True(t, q.hasStarted)

	assert.Equal(t, 3, len(q.success))
}

func TestQueueWithCustomType_Success(t *testing.T) {
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
	assert.Equal(t, 3, len(q.items))

	item, ok := q.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, 1, item.ID)
	assert.Equal(t, "Item 1", item.Name)

	assert.False(t, q.hasFinished)
}

func TestProgressCalculation(t *testing.T) {
	q := NewGenericQueue[int]()

	const itemCount = 100
	for i := 1; i <= itemCount; i++ {
		q.Enqueue(i)
	}

	for i := 0; i < 50; i++ {
		item, ok := q.Dequeue()
		require.True(t, ok)
		q.SetSuccess(item)
	}

	assert.True(t, q.hasStarted)

	progress := q.Progress()
	assert.Equal(t, 50.0, progress.ProgressPct)
	assert.True(t, progress.HasStarted)
	assert.False(t, progress.HasFinished)
	assert.Equal(t, itemCount, progress.TotalItems)
	assert.Equal(t, 50, progress.ProcessedItems)
	assert.Equal(t, 50, progress.SuccessItems)

	// The following are time-dependent and may be flaky in CI environments
	// So we only check that they're set, not their precise values
	assert.NotZero(t, progress.TransferSpeed)
	assert.Equal(t, "items/sec", progress.TransferSpeedUnit)
	assert.NotZero(t, progress.TimeLeft)
	assert.False(t, progress.ETA.IsZero())

	for q.HasRemainingItems() {
		item, ok := q.Dequeue()
		require.True(t, ok)
		q.SetSuccess(item)
	}

	progress = q.Progress()
	assert.Equal(t, 100.0, progress.ProgressPct)
	assert.True(t, progress.HasStarted)
	assert.True(t, progress.HasFinished)
	assert.Equal(t, itemCount, progress.TotalItems)
	assert.Equal(t, itemCount, progress.ProcessedItems)
	assert.Equal(t, itemCount, progress.SuccessItems)
	assert.Zero(t, progress.TimeLeft)
	assert.True(t, progress.ETA.IsZero())
}
