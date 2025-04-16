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

// PreProcess runs a [schema.Pipeline]'s pre-processors on all yet unprocessed
// queue items.
//
// Sorting functions should not be part of pre-processing, as order during
// subsequent concurrent operations cannot be guaranteed. Sorting functions
// should be used as post-processing functions instead.
func (q *EvaluationShareQueue) PreProcess(p schema.Pipeline) bool {
	if moveables, ok := p.PreProcess(q.items); ok {
		q.items = moveables

		return true
	}

	return false
}

// PostProcess runs a [schema.Pipeline]'s post-processors on all successful
// queue items.
//
// Post-processing functions are ideal for sorting, as the order within a share
// is maintained when the [Moveable] are passed to the respective IO queues.
func (q *EvaluationShareQueue) PostProcess(p schema.Pipeline) bool {
	if moveables, ok := p.PostProcess(q.success); ok {
		q.success = moveables

		return true
	}

	return false
}
