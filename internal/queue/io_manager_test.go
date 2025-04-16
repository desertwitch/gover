package queue

import (
	"testing"
	"time"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewIOManager_Success tests the factory function.
func TestNewIOManager_Success(t *testing.T) {
	t.Parallel()

	manager := NewIOManager()
	require.NotNil(t, manager, "NewIOManager() should return a non-nil value")
	require.NotNil(t, manager.GenericManager, "NewIOManager() should initialize the embedded GenericManager")
}

// TestIOManagerEnqueue_Success tests the proper enqueuing of items.
func TestIOManagerEnqueue_Success(t *testing.T) {
	t.Parallel()

	manager := NewIOManager()

	dest1 := schema.NewMockStorage(t)
	dest2 := schema.NewMockStorage(t)

	// Create test Moveables with different target storages
	moveable1 := &schema.Moveable{
		Dest: dest1,
	}

	moveable2 := &schema.Moveable{
		Dest: dest2,
	}

	moveable3 := &schema.Moveable{
		Dest: dest1,
	}

	manager.Enqueue(moveable1, moveable2, moveable3)

	assert.Len(t, manager.queues, 2, "Manager should have 2 queues (one for each unique destination)")

	dest1Queue, exists := manager.queues[dest1]
	require.True(t, exists, "A queue for 'dest1' should exist")
	assert.Len(t, dest1Queue.items, 2, "Queue for 'dest1' should have 2 moveables")

	dest2Queue, exists := manager.queues[dest2]
	require.True(t, exists, "A queue for 'dest2' should exist")
	assert.Len(t, dest2Queue.items, 1, "Queue for 'dest2' should have 1 moveables")

	item, ok := dest1Queue.Dequeue()
	require.True(t, ok, "Should be able to dequeue a moveable from 'dest1' queue")
	assert.Equal(t, moveable1, item, "First moveable from 'dest1' should be moveable1")

	item, ok = dest1Queue.Dequeue()
	require.True(t, ok, "Should be able to dequeue a second moveable from 'dest1' queue")
	assert.Equal(t, moveable3, item, "Second moveable from 'dest1' should be moveable3")

	item, ok = dest2Queue.Dequeue()
	require.True(t, ok, "Should be able to dequeue a moveable from 'dest2' queue")
	assert.Equal(t, moveable2, item, "First moveable from 'dest2' should be moveable2")

	dest1.AssertExpectations(t)
	dest2.AssertExpectations(t)
}

// TestIOManagerProgress_Success tests the progress information.
func TestIOManagerProgress_Success(t *testing.T) {
	t.Parallel()

	manager := NewIOManager()

	dest1 := schema.NewMockStorage(t)
	dest2 := schema.NewMockStorage(t)
	dest3 := schema.NewMockStorage(t)

	moveable1 := &schema.Moveable{
		Dest: dest1,
	}

	moveable2 := &schema.Moveable{
		Dest: dest2,
	}

	moveable3 := &schema.Moveable{
		Dest: dest3,
	}

	manager.Enqueue(moveable1, moveable2, moveable3)

	queues := manager.GetQueues()
	require.Len(t, queues, 3, "should have 3 queues")

	dest1Queue := queues[dest1]
	dest2Queue := queues[dest2]
	dest3Queue := queues[dest3]

	item, ok := dest1Queue.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, item, moveable1)
	dest1Queue.SetSuccess(item)

	item, ok = dest2Queue.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, item, moveable2)
	dest2Queue.SetProcessing(item)

	item, ok = dest3Queue.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, item, moveable3)
	dest3Queue.SetSuccess(item)

	dest1Queue.AddBytesTransfered(1000)
	dest3Queue.AddBytesTransfered(500)

	time.Sleep(10 * time.Millisecond)

	progress := manager.Progress()

	assert.False(t, progress.HasFinished, "Progress should not be finished")
	assert.Equal(t, 1, progress.InProgressItems, "Progress should have one item in progress")
	assert.Equal(t, "bytes/sec", progress.TransferSpeedUnit, "Progress should use 'bytes/sec' as the transfer speed unit")
	assert.Equal(t, 2, progress.SuccessItems, "Progress should report the total completed items across all queues")
	assert.GreaterOrEqual(t, progress.TransferSpeed, 0.0, "Transfer speed should be non-negative when there are pending items and bytes transferred")

	dest1.AssertExpectations(t)
	dest2.AssertExpectations(t)
	dest3.AssertExpectations(t)
}
