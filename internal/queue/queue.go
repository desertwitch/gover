package queue

import (
	"sync"
	"time"
)

// Manager is the principal queue manager implementation.
type Manager struct {
	sync.Mutex
	Mode               int
	EnumerationManager *EnumerationManager
	EvaluationManager  *EvaluationManager
	IOManager          *IOManager
}

// NewManager returns a pointer to a new queue [Manager].
func NewManager() *Manager {
	return &Manager{
		EnumerationManager: NewEnumerationManager(),
		EvaluationManager:  NewEvaluationManager(),
		IOManager:          NewIOManager(),
	}
}

// Progress holds information about the progress of a queue (or manager). It is
// meant to be passed by value.
type Progress struct {
	HasStarted        bool
	HasFinished       bool
	StartTime         time.Time
	FinishTime        time.Time
	ProgressPct       float64
	TotalItems        int
	ProcessedItems    int
	InProgressItems   int
	SuccessItems      int
	SkippedItems      int
	ETA               time.Time
	TimeLeft          time.Duration
	TransferSpeed     float64
	TransferSpeedUnit string
}
