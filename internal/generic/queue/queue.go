package queue

import (
	"sync"
)

type Manager struct {
	sync.Mutex
	Mode               int
	EnumerationManager *EnumerationManager
	EvaluationManager  *EvaluationManager
	IOManager          *IOManager
}

func NewManager() *Manager {
	return &Manager{
		EnumerationManager: NewEnumerationManager(),
		EvaluationManager:  NewEvaluationManager(),
		IOManager:          NewIOManager(),
	}
}
