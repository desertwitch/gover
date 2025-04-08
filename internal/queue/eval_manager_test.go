package queue

import (
	"testing"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewEvaluationManager_Success tests the factory function.
func TestNewEvaluationManager_Success(t *testing.T) {
	t.Parallel()

	manager := NewEvaluationManager()
	require.NotNil(t, manager, "NewEvaluationManager() should return a non-nil value")
	require.NotNil(t, manager.GenericManager, "NewEvaluationManager() should initialize the embedded GenericManager")
}

// TestEvaluationManagerProgress_Success tests progress information.
func TestEvaluationManagerProgress_Success(t *testing.T) {
	t.Parallel()

	manager := NewEvaluationManager()
	progress := manager.Progress()
	assert.Equal(t, "items/sec", progress.TransferSpeedUnit, "Progress should use 'items/sec' as the transfer speed unit")
}

// TestEvaluationManagerEnqueue_Success tests proper enqueuing of items.
func TestEvaluationManagerEnqueue_Success(t *testing.T) {
	t.Parallel()

	manager := NewEvaluationManager()

	share1 := schema.NewMockShare(t)
	share1.EXPECT().GetName().Return("share1")

	share2 := schema.NewMockShare(t)
	share2.EXPECT().GetName().Return("share2")

	item1 := &schema.Moveable{
		Share: share1,
	}

	item2 := &schema.Moveable{
		Share: share2,
	}

	item3 := &schema.Moveable{
		Share: share1,
	}

	manager.Enqueue(item1, item2, item3)

	assert.Len(t, manager.GetQueues(), 2, "Manager should have 2 queues (one for each unique storage)")

	share1Queue, exists := manager.GenericManager.queues["share1"]
	require.True(t, exists, "A queue for 'share1' should exist")
	assert.Len(t, share1Queue.items, 2, "Queue for 'share1' should have 2 items")

	share2Queue, exists := manager.GenericManager.queues["share2"]
	require.True(t, exists, "A queue for 'share2' should exist")
	assert.Len(t, share2Queue.items, 1, "Queue for 'share2' should have 2 items")

	item, ok := share1Queue.Dequeue()
	require.True(t, ok, "Should be able to dequeue a item from 'share1' queue")
	assert.Equal(t, item1, item, "First item from 'share1' should be item1")

	item, ok = share1Queue.Dequeue()
	require.True(t, ok, "Should be able to dequeue a second item from 'share1' queue")
	assert.Equal(t, item3, item, "Second item from 'share1' should be item3")

	item, ok = share2Queue.Dequeue()
	require.True(t, ok, "Should be able to dequeue a item from 'share2' queue")
	assert.Equal(t, item2, item, "First item from 'share2' should be item2")

	share1.AssertExpectations(t)
	share2.AssertExpectations(t)
}
