package queue

import (
	"github.com/desertwitch/gover/internal/schema"
)

type EnumerationTask struct {
	Share    schema.Share
	Source   schema.Storage
	Function func() int
}

func (e *EnumerationTask) Run() int {
	return e.Function()
}

type EnumerationSourceQueue struct {
	*GenericQueue[*EnumerationTask]
}

func NewEnumerationSourceQueue() *EnumerationSourceQueue {
	return &EnumerationSourceQueue{
		GenericQueue: NewGenericQueue[*EnumerationTask](),
	}
}

func (q *EnumerationSourceQueue) Progress() Progress {
	qProgress := q.GenericQueue.Progress()
	qProgress.TransferSpeedUnit = "tasks/sec"

	return qProgress
}
