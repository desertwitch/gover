// Package queue implements structures and routines for organizing tasks and
// items into managed queues. It also provides logic for obtaining progress
// information and metrics on these managed queues.
package queue

import (
	"sync"
	"time"
)

// Manager is the principal queue manager implementation containing:
//   - [queue.EnumerationManager] for enumeration of all [schema.Moveable].
//   - [queue.EvaluationManager] for sorting, allocating and validating them.
//   - [queue.IOManager] for moving all [schema.Moveable] to their destinations.
type Manager struct {
	sync.Mutex

	// EnumerationManager contains all enumeration tasks.
	EnumerationManager *EnumerationManager

	// EvaluationManager contains all [schema.Moveable] candidates.
	EvaluationManager *EvaluationManager

	// IOManager contains all sorted, allocated and validated [schema.Moveable].
	IOManager *IOManager
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
