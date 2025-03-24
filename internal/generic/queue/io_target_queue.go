package queue

import (
	"time"

	"github.com/desertwitch/gover/internal/generic/schema"
)

type IOTargetQueue struct {
	*GenericQueue[*schema.Moveable]
	bytesTransfered uint64
}

func NewIOTargetQueue() *IOTargetQueue {
	return &IOTargetQueue{
		GenericQueue: NewGenericQueue[*schema.Moveable](),
	}
}

func (q *IOTargetQueue) AddBytesTransfered(bytes uint64) {
	q.Lock()
	defer q.Unlock()

	q.bytesTransfered += bytes
}

func (q *IOTargetQueue) Progress() Progress {
	q.RLock()
	defer q.RUnlock()

	isStarted := q.isStarted
	progress := q.head
	totalCount := len(q.items)
	successCount := len(q.success)
	skippedCount := len(q.skipped)

	progress = min(progress, totalCount)

	var progressPct float64
	if totalCount > 0 {
		progressPct = float64(progress) / float64(totalCount) * 100   //nolint:mnd
		progressPct = max(float64(0), min(progressPct, float64(100))) //nolint:mnd
	}

	var eta time.Time
	var timeLeft time.Duration
	var transferSpeed float64
	transferSpeedUnit := "bytes/sec"

	if isStarted && progress > 0 && progress < totalCount {
		elapsed := time.Since(q.startTime)
		itemsPerSec := float64(progress) / elapsed.Seconds()
		bytesPerSec := float64(q.bytesTransfered) / elapsed.Seconds()

		if itemsPerSec > 0 {
			remainingItems := totalCount - progress
			remainingSeconds := float64(remainingItems) / itemsPerSec
			timeLeft = time.Duration(remainingSeconds * float64(time.Second))
			eta = time.Now().Add(timeLeft)
			transferSpeed = bytesPerSec
		}
	}

	return Progress{
		IsStarted:         isStarted,
		StartTime:         q.startTime,
		FinishTime:        q.finishTime,
		ProgressPct:       progressPct,
		TotalItems:        totalCount,
		ProgressItems:     progress,
		SuccessItems:      successCount,
		SkippedItems:      skippedCount,
		ETA:               eta,
		TimeLeft:          timeLeft,
		TransferSpeed:     transferSpeed,
		TransferSpeedUnit: transferSpeedUnit,
	}
}
