package queue

import (
	"time"

	"github.com/desertwitch/gover/internal/schema"
)

// IOManager is a queue manager for IO operations.
// It is used to manage a number of different [IOTargetQueue] that
// are each independent and bucketized by their target storage name.
//
// IOManager embeds a [GenericManager].
// The manager is generally thread-safe and can be accessed concurrently.
//
// Beware that [IOTargetQueue] contained items can only be processed sequentially,
// in order not to operate concurrently within the same destination target storage.
//
// The items contained within [IOTargetQueue] are [schema.Moveable].
type IOManager struct {
	*GenericManager[*schema.Moveable, *IOTargetQueue]
}

// NewIOManager returns a pointer to a new [IOManager].
func NewIOManager() *IOManager {
	return &IOManager{
		GenericManager: NewGenericManager[*schema.Moveable, *IOTargetQueue](),
	}
}

// Enqueue adds [schema.Moveable](s) into the correct [IOTargetQueue], as
// managed by [IOManager], based on their respective target storage name.
func (m *IOManager) Enqueue(items ...*schema.Moveable) {
	for _, item := range items {
		m.GenericManager.Enqueue(item, func(m *schema.Moveable) string {
			return m.Dest.GetName()
		}, NewIOTargetQueue)
	}
}

// Progress returns the [Progress] of the [IOManager].
func (m *IOManager) Progress() Progress {
	mProgress := m.GenericManager.Progress()

	var totalBytesTransferred uint64

	for _, queue := range m.GetQueues() {
		queue.RLock()
		totalBytesTransferred += queue.bytesTransfered
		queue.RUnlock()
	}

	if mProgress.HasStarted && mProgress.ProcessedItems > 0 && mProgress.ProcessedItems < mProgress.TotalItems {
		elapsed := time.Since(mProgress.StartTime)
		bytesPerSec := float64(totalBytesTransferred) / max(elapsed.Seconds(), 1)

		if bytesPerSec > 0 {
			mProgress.TransferSpeed = bytesPerSec
		}
	}

	mProgress.TransferSpeedUnit = "bytes/sec"

	return mProgress
}
