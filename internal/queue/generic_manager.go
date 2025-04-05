package queue

import (
	"sync"
	"time"
)

// GenericQueueType defines methods that a managed queue needs to have.
type GenericQueueType[E comparable] interface {
	Enqueue(items ...E)
	GetSuccessful() []E
	Progress() Progress
}

// GenericManager is a generic queue manager for queues of [GenericQueueType].
type GenericManager[E comparable, T GenericQueueType[E]] struct {
	sync.RWMutex
	queues map[string]T
}

// NewGenericManager returns a pointer to a new [GenericManager].
func NewGenericManager[E comparable, T GenericQueueType[E]]() *GenericManager[E, T] {
	return &GenericManager[E, T]{
		queues: make(map[string]T),
	}
}

// GetSuccessful returns a slice of all queues successfully processed items.
func (m *GenericManager[E, T]) GetSuccessful() []E {
	m.RLock()
	defer m.RUnlock()

	var result []E
	for _, q := range m.queues {
		result = append(result, q.GetSuccessful()...)
	}

	return result
}

// Enqueue bucketizes items into queues according to a getKeyFunc, creating new
// queues as required using a newQueueFunc.
func (m *GenericManager[E, T]) Enqueue(item E, getKeyFunc func(E) string, newQueueFunc func() T) {
	m.Lock()
	defer m.Unlock()

	key := getKeyFunc(item)

	_, exists := m.queues[key]
	if !exists {
		m.queues[key] = newQueueFunc()
	}

	m.queues[key].Enqueue(item)
}

// GetQueues returns a copy of the internal map holding pointers to all managed
// queues.
func (m *GenericManager[E, T]) GetQueues() map[string]T {
	m.RLock()
	defer m.RUnlock()

	if m.queues == nil {
		return nil
	}

	queues := make(map[string]T)

	for k, v := range m.queues {
		queues[k] = v
	}

	return queues
}

// Progress returns the [Progress] for the [GenericManager].
func (m *GenericManager[E, T]) Progress() Progress { //nolint:funlen
	m.RLock()
	defer m.RUnlock()

	if len(m.queues) == 0 {
		return Progress{}
	}

	var totalItems, totalProcessed, totalInProgress, totalSuccess, totalSkipped int

	var anyStarted, allFinished bool
	var earliestStartTime, latestFinishTime time.Time

	for _, queue := range m.queues {
		qProgress := queue.Progress()

		if qProgress.HasStarted {
			if !qProgress.StartTime.IsZero() && (earliestStartTime.IsZero() || qProgress.StartTime.Before(earliestStartTime)) {
				earliestStartTime = qProgress.StartTime
			}
			anyStarted = true
		}

		if !qProgress.FinishTime.IsZero() && (latestFinishTime.IsZero() || qProgress.FinishTime.After(latestFinishTime)) {
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
		allFinished = true
	}

	return Progress{
		HasStarted:        anyStarted,
		HasFinished:       allFinished,
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
