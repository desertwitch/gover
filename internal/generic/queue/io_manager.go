package queue

import (
	"time"

	"github.com/desertwitch/gover/internal/generic/schema"
)

type IOManager struct {
	*GenericManager[*schema.Moveable, *IOTargetQueue]
}

func NewIOManager() *IOManager {
	return &IOManager{
		GenericManager: NewGenericManager[*schema.Moveable, *IOTargetQueue](),
	}
}

func (m *IOManager) Enqueue(items ...*schema.Moveable) {
	for _, item := range items {
		m.GenericManager.Enqueue(item, func(m *schema.Moveable) string {
			return m.Dest.GetName()
		}, NewIOTargetQueue)
	}
}

func (m *IOManager) Progress() Progress {
	mProgress := m.GenericManager.Progress()

	var totalBytesTransferred uint64

	for _, queue := range m.GetQueues() {
		queue.RLock()
		totalBytesTransferred += queue.bytesTransfered
		queue.RUnlock()
	}

	if mProgress.IsStarted && mProgress.ProcessedItems > 0 && mProgress.ProcessedItems < mProgress.TotalItems {
		elapsed := time.Since(mProgress.StartTime)
		bytesPerSec := float64(totalBytesTransferred) / max(elapsed.Seconds(), 1)

		if bytesPerSec > 0 {
			mProgress.TransferSpeed = bytesPerSec
		}
	}

	mProgress.TransferSpeedUnit = "bytes/sec"

	return mProgress
}
