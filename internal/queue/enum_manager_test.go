package queue

import (
	"testing"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewEnumerationManager_Success tests the factory function.
func TestNewEnumerationManager_Success(t *testing.T) {
	t.Parallel()

	manager := NewEnumerationManager()
	require.NotNil(t, manager, "NewEnumerationManager() should return a non-nil value")
	require.NotNil(t, manager.GenericManager, "NewEnumerationManager() should initialize the embedded GenericManager")
}

// TestEnumerationManagerProgress_Success tests progress information.
func TestEnumerationManagerProgress_Success(t *testing.T) {
	t.Parallel()

	manager := NewEnumerationManager()
	progress := manager.Progress()
	assert.Equal(t, "tasks/sec", progress.TransferSpeedUnit, "Progress should use 'tasks/sec' as the transfer speed unit")
}

// TestEnumerationManagerEnqueue_Success tests proper enqueuing of items.
func TestEnumerationManagerEnqueue_Success(t *testing.T) {
	t.Parallel()

	manager := NewEnumerationManager()

	share := schema.NewMockShare(t)

	storage1 := schema.NewMockStorage(t)
	storage2 := schema.NewMockStorage(t)

	task1 := &EnumerationTask{
		Share:    share,
		Source:   storage1,
		Function: func() int { return 1 },
	}

	task2 := &EnumerationTask{
		Share:    share,
		Source:   storage2,
		Function: func() int { return 2 },
	}

	task3 := &EnumerationTask{
		Share:    share,
		Source:   storage1,
		Function: func() int { return 3 },
	}

	manager.Enqueue(task1, task2, task3)

	assert.Len(t, manager.GetQueues(), 2, "Manager should have 2 queues (one for each unique storage)")

	storage1Queue, exists := manager.GenericManager.queues[storage1]
	require.True(t, exists, "A queue for 'storage1' should exist")
	assert.Len(t, storage1Queue.items, 2, "Queue for 'storage1' should have 2 tasks")

	storage2Queue, exists := manager.GenericManager.queues[storage2]
	require.True(t, exists, "A queue for 'storage2' should exist")
	assert.Len(t, storage2Queue.items, 1, "Queue for 'storage2' should have 2 tasks")

	task, ok := storage1Queue.Dequeue()
	require.True(t, ok, "Should be able to dequeue a task from 'storage1' queue")
	assert.Equal(t, task1, task, "First task from 'storage1' should be task1")

	task, ok = storage1Queue.Dequeue()
	require.True(t, ok, "Should be able to dequeue a second task from 'storage1' queue")
	assert.Equal(t, task3, task, "Second task from 'storage1' should be task3")

	task, ok = storage2Queue.Dequeue()
	require.True(t, ok, "Should be able to dequeue a task from 'storage2' queue")
	assert.Equal(t, task2, task, "First task from 'storage2' should be task2")

	storage1.AssertExpectations(t)
	storage2.AssertExpectations(t)
}
