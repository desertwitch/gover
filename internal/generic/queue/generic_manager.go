package queue

import (
	"sync"
	"time"
)

type GenericQueueType[E comparable] interface {
	Enqueue(items ...E)
	GetSuccessful() []E
	Progress() Progress
}

type GenericManager[E comparable, T GenericQueueType[E]] struct {
	sync.RWMutex
	queues map[string]T
}

func NewGenericManager[E comparable, T GenericQueueType[E]]() *GenericManager[E, T] {
	return &GenericManager[E, T]{
		queues: make(map[string]T),
	}
}

func (b *GenericManager[E, T]) GetSuccessful() []E {
	b.RLock()
	defer b.RUnlock()

	var result []E
	for _, q := range b.queues {
		result = append(result, q.GetSuccessful()...)
	}

	return result
}

func (b *GenericManager[E, T]) Enqueue(item E, getKeyFunc func(E) string, newQueueFunc func() T) {
	b.EnqueueSlice([]E{item}, getKeyFunc, newQueueFunc)
}

func (b *GenericManager[E, T]) EnqueueSlice(items []E, getKeyFunc func(E) string, newQueueFunc func() T) {
	b.Lock()
	defer b.Unlock()

	for _, item := range items {
		key := getKeyFunc(item)

		_, exists := b.queues[key]
		if !exists {
			b.queues[key] = newQueueFunc()
		}

		b.queues[key].Enqueue(item)
	}
}

// GetQueues returns a copy of the internal map holding pointers to all queues.
func (b *GenericManager[E, T]) GetQueues() map[string]T {
	b.RLock()
	defer b.RUnlock()

	if b.queues == nil {
		return nil
	}

	queues := make(map[string]T)

	for k, v := range b.queues {
		queues[k] = v
	}

	return queues
}

func (b *GenericManager[E, T]) Progress() Progress {
	b.RLock()
	defer b.RUnlock()

	if len(b.queues) == 0 {
		return Progress{}
	}

	var totalItems, totalProcessed, totalInProgress, totalSuccess, totalSkipped int

	var anyStarted bool
	var earliestStartTime, latestFinishTime time.Time

	for _, queue := range b.queues {
		qProgress := queue.Progress()

		if qProgress.IsStarted {
			if !qProgress.StartTime.IsZero() && (earliestStartTime.IsZero() || qProgress.StartTime.Before(earliestStartTime)) {
				earliestStartTime = qProgress.StartTime
			}
			anyStarted = true
		} else if !qProgress.FinishTime.IsZero() && (latestFinishTime.IsZero() || qProgress.FinishTime.After(latestFinishTime)) {
			latestFinishTime = qProgress.FinishTime
		}

		totalItems += qProgress.TotalItems
		totalProcessed += qProgress.ProcessedItems

		totalInProgress += qProgress.InProgressItems
		totalSuccess += qProgress.SuccessItems
		totalSkipped += qProgress.SkippedItems
	}

	var progressPct float64
	if totalItems > 0 {
		progressPct = float64(totalProcessed) / float64(totalItems) * 100 //nolint:mnd
		progressPct = max(float64(0), min(progressPct, float64(100)))     //nolint:mnd
	}

	var eta time.Time
	var timeLeft time.Duration

	var transferSpeed float64
	transferSpeedUnit := "items/sec"

	if anyStarted && totalProcessed > 0 && totalProcessed < totalItems {
		elapsed := time.Since(earliestStartTime)
		itemsPerSec := float64(totalProcessed) / max(elapsed.Seconds(), 1)

		if itemsPerSec > 0 {
			remainingItems := totalItems - totalProcessed
			remainingSeconds := float64(remainingItems) / itemsPerSec
			timeLeft = time.Duration(remainingSeconds * float64(time.Second))
			eta = time.Now().Add(timeLeft)
			transferSpeed = itemsPerSec
		}
	}

	if anyStarted && (totalProcessed == totalItems) && (totalProcessed > 0) {
		if latestFinishTime.IsZero() {
			latestFinishTime = time.Now()
		}
		anyStarted = false
	}

	return Progress{
		IsStarted:         anyStarted,
		StartTime:         earliestStartTime,
		FinishTime:        latestFinishTime,
		ProgressPct:       progressPct,
		TotalItems:        totalItems,
		ProcessedItems:    totalProcessed,
		InProgressItems:   totalInProgress,
		SuccessItems:      totalSuccess,
		SkippedItems:      totalSkipped,
		ETA:               eta,
		TimeLeft:          timeLeft,
		TransferSpeed:     transferSpeed,
		TransferSpeedUnit: transferSpeedUnit,
	}
}
