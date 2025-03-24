package queue

type EnumerationManager struct {
	*GenericManager[*EnumerationTask, *EnumerationSourceQueue]
}

func NewEnumerationManager() *EnumerationManager {
	return &EnumerationManager{
		GenericManager: NewGenericManager[*EnumerationTask, *EnumerationSourceQueue](),
	}
}
