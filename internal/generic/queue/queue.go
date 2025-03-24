package queue

import (
	"sync"
	"time"
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
