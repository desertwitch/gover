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
	qProgress := q.GenericQueue.Progress()

	if qProgress.IsStarted && qProgress.ProcessedItems > 0 && qProgress.ProcessedItems < qProgress.TotalItems {
		elapsed := time.Since(qProgress.StartTime)

		q.RLock()
		bytesPerSec := float64(q.bytesTransfered) / max(elapsed.Seconds(), 1)
		q.RUnlock()

		if bytesPerSec > 0 {
			qProgress.TransferSpeed = bytesPerSec
		}
	}

	qProgress.TransferSpeedUnit = "bytes/sec"

	return qProgress
}
