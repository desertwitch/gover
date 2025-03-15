package queue

import (
	"sync"
)

type Manager struct {
	sync.Mutex
	Mode             int
	EnumerationQueue *EnumerationQueue
	IOManager        *IOManager
}

func NewManager() *Manager {
	return &Manager{
		EnumerationQueue: NewEnumerationQueue(),
		IOManager:        NewIOManager(),
	}
}
