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

func (b *IOManager) Progress() Progress {
	mProgress := b.GenericManager.Progress()

	var totalBytesTransferred uint64

	for _, queue := range b.GetQueues() {
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
