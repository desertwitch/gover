package unraid

import (
	"fmt"
	"strings"
)

// findPool dereferences a textual pool name into a pool pointer.
func findPool(pools map[string]*Pool, poolName string) (*Pool, error) {
	if poolName == "" {
		return nil, nil
	}
	if pool, exists := pools[poolName]; exists {
		return pool, nil
	}

	return nil, fmt.Errorf("configured pool %s does not exist", poolName)
}

// findDisks dereferences a list of textual disk names into a map of disk pointers.
func findDisks(disks map[string]*Disk, diskNames string) (map[string]*Disk, error) {
	if diskNames == "" {
		return nil, nil
	}

	diskList := strings.Split(diskNames, ",")
	foundDisks := make(map[string]*Disk)

	for _, name := range diskList {
		if disk, exists := disks[name]; exists {
			foundDisks[name] = disk
		} else {
			return nil, fmt.Errorf("configured disk %s does not exist", name)
		}
	}

	return foundDisks, nil
}
