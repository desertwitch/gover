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

func (m *EvaluationManager) Enqueue(items ...*schema.Moveable) {
	for _, item := range items {
		m.GenericManager.Enqueue(item, func(m *schema.Moveable) string {
			return m.Share.GetName()
		}, NewEvaluationShareQueue)
	}
}
