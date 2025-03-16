package queue

import (
	"sync"
)

type Manager struct {
	sync.Mutex
	Mode               int
	EnumerationManager *EnumerationManager
	IOManager          *IOManager
}

func NewManager() *Manager {
	return &Manager{
		EnumerationManager: NewEnumerationManager(),
		IOManager:          NewIOManager(),
	}
}
