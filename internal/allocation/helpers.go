package allocation

import "github.com/desertwitch/gover/internal/unraid"

func (a *Handler) getAlreadyAllocated(disk *unraid.Disk) uint64 {
	a.RLock()
	defer a.RUnlock()

	return a.alreadyAllocated[disk]
}

func (a *Handler) addAlreadyAllocated(disk *unraid.Disk, size uint64) {
	a.Lock()
	defer a.Unlock()

	a.alreadyAllocated[disk] += size
}
