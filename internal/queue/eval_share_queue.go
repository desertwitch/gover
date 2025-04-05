package queue

import (
	"github.com/desertwitch/gover/internal/schema"
)

// EvaluationShareQueue is a queue where items of a common share name were
// previously enqueued and aggregated by their [EvaluationManager].
//
// EvaluationShareQueue embeds a [GenericQueue]. It is thread-safe and can both
// be accessed and processed concurrently.
//
// The items contained within [EvaluationShareQueue] are [schema.Moveable].
type EvaluationShareQueue struct {
	*GenericQueue[*schema.Moveable]
}

// NewEvaluationShareQueue returns a pointer to a new [EvaluationShareQueue].
// This method is generally only called from the respective [EvaluationManager].
func NewEvaluationShareQueue() *EvaluationShareQueue {
	return &EvaluationShareQueue{
		GenericQueue: NewGenericQueue[*schema.Moveable](),
	}
}

// Progress returns the [Progress] of the [EvaluationShareQueue].
func (q *EvaluationShareQueue) Progress() Progress {
	qProgress := q.GenericQueue.Progress()
	qProgress.TransferSpeedUnit = "items/sec"

	return qProgress
}
