package queue

import (
	"context"
	"time"

	"github.com/desertwitch/gover/internal/schema"
)

// IOTargetQueue is a queue where items of a common target storage name were
// previously enqueued and aggregated by their [IOManager].
//
// IOTargetQueue embeds a [GenericQueue].
//
// Beware that [IOTargetQueue] contained items can only be processed
// sequentially, in order not to operate concurrently within the same
// destination target storage.
//
// The items contained within [IOTargetQueue] are [schema.Moveable].
type IOTargetQueue struct {
	*GenericQueue[*schema.Moveable]

	// bytesTransfered is the amount of bytes transferred for the
	// [IOTargetQueue].
	bytesTransfered uint64
}

// NewIOTargetQueue returns a pointer to a new [IOTargetQueue]. This method is
// generally only called from the respective [IOManager].
func NewIOTargetQueue() *IOTargetQueue {
	return &IOTargetQueue{
		GenericQueue: NewGenericQueue[*schema.Moveable](),
	}
}

// DequeueAndProcessConc is unsupported by [IOTargetQueue] and will result in a
// panic when used.
func (q *IOTargetQueue) DequeueAndProcessConc(ctx context.Context, maxWorkers int, processFunc func(*schema.Moveable) int) error { //nolint:revive
	panic("An IOTargetQueue cannot be processed concurrently.")
}

// AddBytesTransfered adds given transferred bytes to the total amount
// transferred for that [IOTargetQueue].
func (q *IOTargetQueue) AddBytesTransfered(bytes uint64) {
	q.Lock()
	defer q.Unlock()

	q.bytesTransfered += bytes
}

// Progress returns the [Progress] of the [IOTargetQueue].
func (q *IOTargetQueue) Progress() Progress {
	qProgress := q.GenericQueue.Progress()

	if qProgress.HasStarted && qProgress.ProcessedItems > 0 && qProgress.ProcessedItems < qProgress.TotalItems {
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
