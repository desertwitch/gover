package allocation

import (
	"path/filepath"
	"strings"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/unraid"
)

func (a *Handler) getAllocatedsForSubpath(subPath string) map[string]*unraid.Disk {
	a.RLock()
	defer a.RUnlock()

	found := make(map[string]*unraid.Disk)

	for _, allocInfo := range a.alreadyAllocated {
		checkPath := filepath.Join(allocInfo.sourceBase, subPath)
		if strings.HasPrefix(allocInfo.sourcePath, checkPath) {
			found[allocInfo.allocatedDisk.Name] = allocInfo.allocatedDisk
		}
	}

	return found
}

func (a *Handler) getAllocatedSpace(disk *unraid.Disk) uint64 {
	a.RLock()
	defer a.RUnlock()

	return a.alreadyAllocatedSpace[disk]
}

func (a *Handler) addAllocated(m *filesystem.Moveable, dst *unraid.Disk) {
	a.Lock()
	defer a.Unlock()

	allocInfo := &allocInfo{
		sourcePath:    m.SourcePath,
		sourceBase:    m.Source.GetFSPath(),
		allocatedDisk: dst,
	}

	a.alreadyAllocated[m] = allocInfo
}

func (a *Handler) addAllocatedSpace(disk *unraid.Disk, size uint64) {
	a.Lock()
	defer a.Unlock()

	a.alreadyAllocatedSpace[disk] += size
}
