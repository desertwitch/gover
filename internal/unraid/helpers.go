package unraid

import (
	"fmt"
	"strings"
)

// findPool cross-references a textual pool name with a map (map[diskName]*Pool)
// of already known [Pool]. A pointer to the [Pool] that was found is returned.
//
// If the poolName is empty, both the pointer and error are returend as nil. If
// the poolName is non-empty, but cannot be found among the knownPools,
// [ErrConfPoolNotFound] is returned.
func findPool(poolName string, knownPools map[string]*Pool) (*Pool, error) {
	if poolName == "" {
		return nil, nil //nolint: nilnil
	}
	if pool, exists := knownPools[poolName]; exists {
		return pool, nil
	}

	return nil, fmt.Errorf("(unraid-findpool) %w: %s", ErrConfPoolNotFound, poolName)
}

// findDisks cross-references a comma-separated list of disk names with a map
// (map[diskName]*Disk) of already known [Disk]. A map (map[diskName]*Disk) of
// all matching [Disk] is returned.
//
// If the diskNames is empty, both the map and error are returend as nil. If the
// diskNames is non-empty, but cannot be found among the knownDisks,
// [ErrConfDiskNotFound] is returned.
func findDisks(diskNames string, knownDisks map[string]*Disk) (map[string]*Disk, error) {
	if diskNames == "" {
		return nil, nil //nolint: nilnil
	}

	foundDisks := make(map[string]*Disk)

	diskList := strings.SplitSeq(diskNames, ",")
	for name := range diskList {
		if disk, exists := knownDisks[name]; exists {
			foundDisks[name] = disk
		} else {
			return nil, fmt.Errorf("(unraid-finddisk) %w: %s", ErrConfDiskNotFound, name)
		}
	}

	return foundDisks, nil
}
