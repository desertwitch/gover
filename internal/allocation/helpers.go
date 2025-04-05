package allocation

import (
	"path/filepath"
	"strings"

	"github.com/desertwitch/gover/internal/schema"
)

// getAllocatedsForSubpath is a thread-safe method to check if a certain subpath
// is already allocated to one or multiple [schema.Disk] and returns those in a
// map (map[diskName]schema.Disk) for later use in e.g. split-level allocation
// decisions.
func (a *Handler) getAllocatedsForSubpath(subPath string) map[string]schema.Disk {
	a.RLock()
	defer a.RUnlock()

	found := make(map[string]schema.Disk)

	for _, allocInfo := range a.alreadyAllocated {
		checkPath := filepath.Join(allocInfo.sourceBase, subPath)
		if strings.HasPrefix(allocInfo.sourcePath, checkPath) {
			found[allocInfo.allocatedDisk.GetName()] = allocInfo.allocatedDisk
		}
	}

	return found
}

// getAllocatedSpace is a thread-safe method to retrieve the already allocated
// space for a specific [schema.Disk].
func (a *Handler) getAllocatedSpace(disk schema.Disk) uint64 {
	a.RLock()
	defer a.RUnlock()

	return a.alreadyAllocatedSpace[disk.GetName()]
}

// addAllocated is a thread-safe method to add [allocInfo] for a specific
// [schema.Moveable] to the [Handler] map of total [schema.Moveable]
// allocations.
func (a *Handler) addAllocated(m *schema.Moveable, dst schema.Disk) {
	a.Lock()
	defer a.Unlock()

	allocInfo := allocInfo{
		sourcePath:    m.SourcePath,
		sourceBase:    m.Source.GetFSPath(),
		allocatedDisk: dst,
	}

	a.alreadyAllocated[m] = allocInfo
}

// addAllocatedSpace is a thread-safe method to add allocation size for a
// [schema.Disk] to the [Handler] map of total allocated space per
// [schema.Disk].
func (a *Handler) addAllocatedSpace(disk schema.Disk, size uint64) {
	a.Lock()
	defer a.Unlock()

	a.alreadyAllocatedSpace[disk.GetName()] += size
}
