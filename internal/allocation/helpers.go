package allocation

import (
	"path/filepath"
	"strings"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/storage"
)

func (a *Handler) getAllocatedsForSubpath(subPath string) map[string]storage.Disk {
	a.RLock()
	defer a.RUnlock()

	found := make(map[string]storage.Disk)

	for _, allocInfo := range a.alreadyAllocated {
		checkPath := filepath.Join(allocInfo.sourceBase, subPath)
		if strings.HasPrefix(allocInfo.sourcePath, checkPath) {
			found[allocInfo.allocatedDisk.GetName()] = allocInfo.allocatedDisk
		}
	}

	return found
}

func (a *Handler) getAllocatedSpace(disk storage.Disk) uint64 {
	a.RLock()
	defer a.RUnlock()

	return a.alreadyAllocatedSpace[disk.GetName()]
}

func (a *Handler) addAllocated(m *filesystem.Moveable, dst storage.Disk) {
	a.Lock()
	defer a.Unlock()

	allocInfo := &allocInfo{
		sourcePath:    m.SourcePath,
		sourceBase:    m.Source.GetFSPath(),
		allocatedDisk: dst,
	}

	a.alreadyAllocated[m] = allocInfo
}

func (a *Handler) addAllocatedSpace(disk storage.Disk, size uint64) {
	a.Lock()
	defer a.Unlock()

	a.alreadyAllocatedSpace[disk.GetName()] += size
}
