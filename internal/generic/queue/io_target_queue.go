package queue

import (
	"github.com/desertwitch/gover/internal/generic/schema"
)

type IOTargetQueue struct {
	*GenericQueue[*schema.Moveable]
}

func NewIOTargetQueue() *IOTargetQueue {
	return &IOTargetQueue{
		GenericQueue: NewGenericQueue[*schema.Moveable](),
	}
}
