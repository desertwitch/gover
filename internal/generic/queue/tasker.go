package queue

import (
	"context"
	"sync"
)

type TaskManager struct {
	sync.Mutex
	Tasks []func()
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		Tasks: []func(){},
	}
}

func (t *TaskManager) Add(taskedFunc func()) {
	t.Lock()
	defer t.Unlock()

	t.Tasks = append(t.Tasks, taskedFunc)
}

func (t *TaskManager) Launch(ctx context.Context) {
	t.Lock()
	defer t.Unlock()

	for _, task := range t.Tasks {
		task()
	}
}

func (t *TaskManager) LaunchConcAndWait(ctx context.Context, maxWorkers int) error {
	t.Lock()
	defer t.Unlock()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxWorkers)

	for _, task := range t.Tasks {
		select {
		case <-ctx.Done():
			wg.Wait()

			return ctx.Err()
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

	return nil
}
