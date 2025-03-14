package queue

type Manager struct {
	EnumerationManager *EnumerationManager
	IOManager          *IOManager
}

func NewManager() *Manager {
	return &Manager{
		EnumerationManager: NewEnumerationManager(),
		IOManager:          NewIOManager(),
	}
}
