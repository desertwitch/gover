package queue

import (
	"context"
	"fmt"
	"sync"
)

// TaskManager is a simple task manager for delayed function execution.
type TaskManager struct {
	sync.Mutex
	Tasks []func()
}

// NewTaskManager returns a pointer to a new [TaskManager].
func NewTaskManager() *TaskManager {
	return &TaskManager{
		Tasks: []func(){},
	}
}

// Add adds a new taskedFunc to the [TaskManager].
// Functions with parameters can be added by invoking a parameterized function
// that immediately returns a func(), capturing any parameters in the closure.
func (t *TaskManager) Add(taskedFunc func()) {
	t.Lock()
	defer t.Unlock()

	t.Tasks = append(t.Tasks, taskedFunc)
}

// Launch sequentially launches the functions stored in a [TaskManager].
// An error is only returned in case of a mid-flight context cancellation.
func (t *TaskManager) Launch(ctx context.Context) error {
	t.Lock()
	defer t.Unlock()

	for _, task := range t.Tasks {
		if ctx.Err() != nil {
			break
		}

		task()
	}

	if ctx.Err() != nil {
		return fmt.Errorf("(queue-tasker) %w", ctx.Err())
	}

	return nil
}

// LaunchConcAndWait concurrently launches the functions stored in a [TaskManager].
// An error is only returned in case of a mid-flight context cancellation.
//
// It is the responsibility of the taskedFunc to ensure thread-safety for anything happening
// inside the taskedFunc, with the [TaskManager] only guaranteeing thread-safety for itself.
func (t *TaskManager) LaunchConcAndWait(ctx context.Context, maxWorkers int) error {
	t.Lock()
	defer t.Unlock()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxWorkers)

	for _, task := range t.Tasks {
		select {
		case <-ctx.Done():
			wg.Wait()

			return fmt.Errorf("(queue-tasker-conc) %w", ctx.Err())
		case semaphore <- struct{}{}:
		}

		wg.Add(1)
		go func(task func()) {
			defer wg.Done()
			defer func() { <-semaphore }()

			task()
		}(task)
	}

	wg.Wait()

	if ctx.Err() != nil {
		return fmt.Errorf("(queue-tasker-conc) %w", ctx.Err())
	}

	return nil
}
