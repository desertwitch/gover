package queue

import (
	"sync"
	"testing"
	"time"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewIOTargetQueue_Success test the factory function.
func TestNewIOTargetQueue_Success(t *testing.T) {
	t.Parallel()

	queue := NewIOTargetQueue()

	require.NotNil(t, queue, "NewIOTargetQueue() should return a non-nil value")
	require.NotNil(t, queue.GenericQueue, "NewIOTargetQueue() should initialize the embedded GenericQueue")
	assert.Equal(t, uint64(0), queue.bytesTransfered, "NewIOTargetQueue() should initialize bytesTransfered to 0")
}

// TestIOTargetQueueDequeueAndProcessConc_Fail_Panic tests the non-supported
// concurrent processing.
func TestIOTargetQueueDequeueAndProcessConc_Fail_Panic(t *testing.T) {
	t.Parallel()

	queue := NewIOTargetQueue()

	assert.Panics(t, func() {
		ctx := t.Context()
		_ = queue.DequeueAndProcessConc(ctx, 2, func(m *schema.Moveable) int { return 0 })
	}, "DequeueAndProcessConc should panic when called on IOTargetQueue")
}

// TestIOTargetQueueAddBytesTransfered_Success tests adding transferred bytes.
func TestIOTargetQueueAddBytesTransfered_Success(t *testing.T) {
	t.Parallel()

	queue := NewIOTargetQueue()

	queue.AddBytesTransfered(100)
	assert.Equal(t, uint64(100), queue.bytesTransfered, "AddBytesTransfered should add bytes to the total")

	queue.AddBytesTransfered(50)
	assert.Equal(t, uint64(150), queue.bytesTransfered, "AddBytesTransfered should accumulate bytes correctly")

	queue.AddBytesTransfered(0)
	assert.Equal(t, uint64(150), queue.bytesTransfered, "AddBytesTransfered with 0 should not change the total")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 50 {
			queue.AddBytesTransfered(uint64(1))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range 50 {
			queue.AddBytesTransfered(uint64(1))
		}
	}()

	wg.Wait()

	assert.Equal(t, uint64(250), queue.bytesTransfered, "AddBytesTransfered should sum properly under concurrence.")
}

// TestIOTargetQueueProgress_Success tests progress reporting.
func TestIOTargetQueueProgress_Success(t *testing.T) {
	t.Parallel()

	queue := NewIOTargetQueue()
	moveable := &schema.Moveable{}

	queue.Enqueue(moveable)
	item, ok := queue.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, item, moveable)
	queue.SetSuccess(item)

	queue.AddBytesTransfered(1000)
	time.Sleep(10 * time.Millisecond)

	progress := queue.Progress()
	assert.Equal(t, "bytes/sec", progress.TransferSpeedUnit, "Progress should use 'bytes/sec' as the transfer speed unit")
	assert.Equal(t, 1, progress.SuccessItems, "Progress should report completed items correctly")

	queue = NewIOTargetQueue()

	queue.Enqueue(moveable)
	queue.Enqueue(moveable)

	item, ok = queue.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, item, moveable)
	queue.SetSuccess(item)

	queue.AddBytesTransfered(1000)

	progress = queue.Progress()

	assert.Equal(t, "bytes/sec", progress.TransferSpeedUnit, "Progress should use 'bytes/sec' as the transfer speed unit")
	assert.GreaterOrEqual(t, progress.TransferSpeed, 0.0, "Transfer speed should be non-negative when active transfers exist")
}
