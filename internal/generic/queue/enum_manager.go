package queue

type EnumerationManager struct {
	*GenericManager[*EnumerationTask, *EnumerationSourceQueue]
}

func NewEnumerationManager() *EnumerationManager {
	return &EnumerationManager{
		GenericManager: NewGenericManager[*EnumerationTask, *EnumerationSourceQueue](),
	}
}

func (m *EnumerationManager) Progress() Progress {
	mProgress := m.GenericManager.Progress()
	mProgress.TransferSpeedUnit = "tasks/sec"

	return mProgress
}
