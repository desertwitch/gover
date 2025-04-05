package queue

import (
	"github.com/desertwitch/gover/internal/schema"
)

// EvaluationManager is a queue manager for evaluation operations. It is used to
// manage a number of different [EvaluationShareQueue] that are each independent
// and bucketized by using their share name.
//
// EvaluationManager embeds a [GenericManager]. It is thread-safe and can both
// be accessed and processed concurrently.
//
// The items contained within [EvaluationShareQueue] are [schema.Moveable].
type EvaluationManager struct {
	*GenericManager[*schema.Moveable, *EvaluationShareQueue]
}

// NewEvaluationManager returns a pointer to a new [EvaluationManager].
func NewEvaluationManager() *EvaluationManager {
	return &EvaluationManager{
		GenericManager: NewGenericManager[*schema.Moveable, *EvaluationShareQueue](),
	}
}

// Progress returns the [Progress] of the [EvaluationManager].
func (m *EvaluationManager) Progress() Progress {
	mProgress := m.GenericManager.Progress()
	mProgress.TransferSpeedUnit = "tasks/sec"

	return mProgress
}

// Enqueue adds [schema.Moveable](s) into the correct [EvaluationShareQueue], as
// managed by [EvaluationManager], based on their respective shares names.
func (m *EvaluationManager) Enqueue(items ...*schema.Moveable) {
	for _, item := range items {
		m.GenericManager.Enqueue(item, func(m *schema.Moveable) string {
			return m.Share.GetName()
		}, NewEvaluationShareQueue)
	}
}
