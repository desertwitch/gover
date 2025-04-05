package queue

import (
	"github.com/desertwitch/gover/internal/schema"
)

// EnumerationTask is an enumeration task for later execution.
type EnumerationTask struct {
	Share    schema.Share
	Source   schema.Storage
	Function func() int
}

// Run executes the stored enumeration function of an [EnumerationTask].
func (e *EnumerationTask) Run() int {
	return e.Function()
}

// EnumerationSourceQueue is a queue where items of a common source storage
// name were previously enqueued and aggregated by their [EnumerationManager].
//
// EnumerationSourceQueue embeds a [GenericQueue].
// It is thread-safe and can both be accessed and processed concurrently.
//
// The items contained within [EnumerationSourceQueue] are [EnumerationTask].
type EnumerationSourceQueue struct {
	*GenericQueue[*EnumerationTask]
}

// NewEnumerationSourceQueue returns a pointer to a new [EnumerationSourceQueue].
// This method is generally only called from the respective [EnumerationManager].
func NewEnumerationSourceQueue() *EnumerationSourceQueue {
	return &EnumerationSourceQueue{
		GenericQueue: NewGenericQueue[*EnumerationTask](),
	}
}

// Progress returns the [Progress] of the [EnumerationSourceQueue].
func (q *EnumerationSourceQueue) Progress() Progress {
	qProgress := q.GenericQueue.Progress()
	qProgress.TransferSpeedUnit = "tasks/sec"

	return qProgress
}
