package queue

import (
	"github.com/desertwitch/gover/internal/generic/schema"
)

type EvaluationManager struct {
	*GenericManager[*schema.Moveable, *EvaluationShareQueue]
}

func NewEvaluationManager() *EvaluationManager {
	return &EvaluationManager{
		GenericManager: NewGenericManager[*schema.Moveable, *EvaluationShareQueue](),
	}
}
