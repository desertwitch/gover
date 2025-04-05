package queue

// EnumerationManager is a queue manager for enumeration operations. It is used
// to manage a number of different [EnumerationSourceQueue] that are each
// independent and bucketized by their source storage name.
//
// EnumerationManager embeds a [GenericManager]. It is thread-safe and can both
// be accessed and processed concurrently.
//
// The items contained within [EnumerationSourceQueue] are [EnumerationTask].
type EnumerationManager struct {
	*GenericManager[*EnumerationTask, *EnumerationSourceQueue]
}

// NewEnumerationManager returns a pointer to a new [EnumerationManager].
func NewEnumerationManager() *EnumerationManager {
	return &EnumerationManager{
		GenericManager: NewGenericManager[*EnumerationTask, *EnumerationSourceQueue](),
	}
}

// Progress returns the [Progress] of the [EnumerationManager].
func (m *EnumerationManager) Progress() Progress {
	mProgress := m.GenericManager.Progress()
	mProgress.TransferSpeedUnit = "tasks/sec"

	return mProgress
}

// Enqueue adds [EnumerationTask](s) into the correct [EnumerationSourceQueue],
// as managed by [EnumerationManager], based on their respective source storage
// name.
func (m *EnumerationManager) Enqueue(items ...*EnumerationTask) {
	for _, item := range items {
		m.GenericManager.Enqueue(item, func(et *EnumerationTask) string {
			return et.Source.GetName()
		}, NewEnumerationSourceQueue)
	}
}
