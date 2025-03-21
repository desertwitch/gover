package queue

import (
	"github.com/desertwitch/gover/internal/generic/schema"
)

type EvaluationShareQueue struct {
	*GenericQueue[*schema.Moveable]
}

func NewEvaluationShareQueue() *EvaluationShareQueue {
	return &EvaluationShareQueue{
		GenericQueue: NewGenericQueue[*schema.Moveable](),
	}
}
