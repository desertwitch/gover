package unraid

import (
	"fmt"
	"strings"
)

// findPool dereferences a textual pool name into a pool pointer.
func findPool(pools map[string]*Pool, poolName string) (*Pool, error) {
	if poolName == "" {
		return nil, nil //nolint: nilnil
	}
	if pool, exists := pools[poolName]; exists {
		return pool, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrConfPoolNotFound, poolName)
}

// findDisks dereferences a list of textual disk names into a map of disk pointers.
func findDisks(disks map[string]*Disk, diskNames string) (map[string]*Disk, error) {
	if diskNames == "" {
		return nil, nil //nolint: nilnil
	}

	diskList := strings.Split(diskNames, ",")
	foundDisks := make(map[string]*Disk)

	for _, name := range diskList {
		if disk, exists := disks[name]; exists {
			foundDisks[name] = disk
		} else {
			return nil, fmt.Errorf("%w: %s", ErrConfDiskNotFound, name)
		}
	}

	return foundDisks, nil
}
