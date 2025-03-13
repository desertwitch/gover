package util

import (
	"context"
	"runtime"
	"sync"
)

func ConcurrentFilterSlice[T any](ctx context.Context, items []T, filterFunc func(T) bool) ([]T, error) {
	var wg sync.WaitGroup

	ch := make(chan T, len(items))

	wg.Add(1)
	go func() {
		defer wg.Done()

		maxWorkers := runtime.NumCPU()
		semaphore := make(chan struct{}, maxWorkers)

		for _, item := range items {
			select {
			case <-ctx.Done():
				return
			case semaphore <- struct{}{}:
			}

			wg.Add(1)
			go func(item T) {
				defer wg.Done()
				defer func() { <-semaphore }()

				if filterFunc(item) {
					select {
					case <-ctx.Done():
						return
					case ch <- item:
					}
				}
			}(item)
		}
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	filtered := []T{}
	for item := range ch {
		filtered = append(filtered, item)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return filtered, nil
}
