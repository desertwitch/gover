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

// TestNewTaskManager_Success tests the factory function.
func TestNewTaskManager_Success(t *testing.T) {
	t.Parallel()
	tm := NewTaskManager()
	require.NotNil(t, tm, "NewTaskManager() should return a non-nil value")
	assert.NotNil(t, tm.Tasks, "NewTaskManager() should initialize Tasks slice")
	assert.Empty(t, tm.Tasks, "NewTaskManager() should initialize an empty Tasks slice")
}

// TestTaskManagerAdd_Success tests adding tasks.
func TestTaskManagerAdd_Success(t *testing.T) {
	t.Parallel()

	tm := NewTaskManager()

	counter := 0
	tm.Add(func() { counter++ })
	tm.Add(func() { counter += 2 })

	assert.Len(t, tm.Tasks, 2, "Add should append tasks to the Tasks slice")
	assert.Empty(t, counter, "Tasks should not be executed when added")
}

// TestTaskManagerLaunch_Success tests sequential processing the tasked
// functions.
func TestTaskManagerLaunch_Success(t *testing.T) {
	t.Parallel()

	tm := NewTaskManager()
	err := tm.Launch(t.Context())
	require.NoError(t, err, "Launch should not return an error with empty task list")

	executionOrder := []int{}
	tm.Add(func() { executionOrder = append(executionOrder, 1) })
	tm.Add(func() { executionOrder = append(executionOrder, 2) })
	tm.Add(func() { executionOrder = append(executionOrder, 3) })

	err = tm.Launch(t.Context())
	require.NoError(t, err, "Launch should not return an error when all tasks complete")
	assert.Equal(t, []int{1, 2, 3}, executionOrder, "Tasks should execute sequentially in the order they were added")
}

// TestTaskManagerLaunch_Fail_CtxCancel tests in-flight context cancellation
// during sequential processing.
func TestTaskManagerLaunch_Fail_CtxCancel(t *testing.T) {
	t.Parallel()

	cancelCtx, cancel := context.WithCancel(t.Context())
	tm := NewTaskManager()

	executed := false
	canceled := false

	tm.Add(func() { executed = true })

	tm.Add(func() {
		cancel()
		canceled = true
	})

	shouldNotExecute := false
	tm.Add(func() { shouldNotExecute = true })

	err := tm.Launch(cancelCtx)
	require.ErrorIs(t, err, context.Canceled, "Launch should return an error when context is canceled")
	assert.True(t, executed, "Tasks before cancellation should execute")
	assert.True(t, canceled, "The task that performs cancellation should execute")
	assert.False(t, shouldNotExecute, "Tasks after cancellation should not execute")
}

// TestTaskManagerLaunchConcAndWait_Success tests concurrent task processing.
func TestTaskManagerLaunchConcAndWait_Success(t *testing.T) {
	t.Parallel()

	tm := NewTaskManager()

	err := tm.LaunchConcAndWait(t.Context(), 2)
	require.NoError(t, err, "LaunchConcAndWait should not return an error with empty task list")

	var counter atomic.Int32
	var wg sync.WaitGroup

	wg.Add(1)
	tm.Add(func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		counter.Add(1)
	})

	wg.Add(1)
	tm.Add(func() {
		defer wg.Done()
		time.Sleep(30 * time.Millisecond)
		counter.Add(1)
	})

	wg.Add(1)
	tm.Add(func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		counter.Add(1)
	})

	err = tm.LaunchConcAndWait(t.Context(), 3)
	wg.Wait()

	require.NoError(t, err, "LaunchConcAndWait should not return an error when all tasks complete")
	assert.Equal(t, 3, int(counter.Load()), "All tasks should have executed")
}

// TestTaskManagerLaunchConcAndWait_Success_WorkerLimit tests the worker limit
// is respected.
func TestTaskManagerLaunchConcAndWait_Success_WorkerLimit(t *testing.T) {
	t.Parallel()

	tm := NewTaskManager()
	barrier := make(chan struct{})

	maxWorkers := 2
	var inFlightCount atomic.Int32
	var taskWg sync.WaitGroup

	taskWg.Add(50)
	for range 50 {
		tm.Add(func() {
			defer taskWg.Done()
			<-barrier

			inFlightCount.Add(1)
			defer inFlightCount.Add(-1)

			require.LessOrEqual(t, int(inFlightCount.Load()), maxWorkers,
				"Number of concurrently executing tasks should not exceed maxWorkers")

			time.Sleep(10 * time.Millisecond)
		})
	}

	go func() {
		close(barrier)
	}()

	err := tm.LaunchConcAndWait(t.Context(), maxWorkers)
	require.NoError(t, err, "LaunchConcAndWait should not return an error with limited workers")
}

// TestTaskManagerLaunchConcAndWait_Fail_CtxCancel tests in-flight context
// cancellation during concurrent processing.
func TestTaskManagerLaunchConcAndWait_Fail_CtxCancel(t *testing.T) {
	t.Parallel()

	cancelCtx, cancel := context.WithCancel(t.Context())
	tm := NewTaskManager()

	var executed atomic.Int32

	for range 10 {
		tm.Add(func() {
			time.Sleep(100 * time.Millisecond)
			executed.Add(1)
		})
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := tm.LaunchConcAndWait(cancelCtx, 2)
	require.ErrorIs(t, err, context.Canceled, "LaunchConcAndWait should return an error when context is canceled")
	assert.Less(t, int(executed.Load()), 10, "Not all tasks should execute when context is canceled")
}
