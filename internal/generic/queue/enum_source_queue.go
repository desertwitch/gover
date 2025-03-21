package queue

import (
	"github.com/desertwitch/gover/internal/generic/schema"
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
