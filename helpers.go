package main

import (
	"fmt"
	"strconv"
	"strings"
)

// getConfigValue returns an element of a string map or "" if not existing
func getConfigValue(envMap map[string]string, key string) string {
	if value, exists := envMap[key]; exists {
		return value
	}
	return ""
}

// findPool dereferences a textual pool name into an established pool pointer
func findPool(pools map[string]*UnraidPool, poolName string) (*UnraidPool, error) {
	if poolName == "" {
		return nil, nil
	}
	if pool, exists := pools[poolName]; exists {
		return pool, nil
	}
	return nil, fmt.Errorf("configured pool %s not found in mounted pools", poolName)
}

// findDisks dereferences a list of textual disk names into a map of established disk pointers
func findDisks(disks map[string]*UnraidDisk, diskNames string) (map[string]*UnraidDisk, error) {
	if diskNames == "" {
		return nil, nil
	}

	diskList := strings.Split(diskNames, ",")
	foundDisks := make(map[string]*UnraidDisk)

	for _, name := range diskList {
		if disk, exists := disks[name]; exists {
			foundDisks[name] = disk
		} else {
			return nil, fmt.Errorf("configured disk %s not found in mounted disks", name)
		}
	}

	return foundDisks, nil
}

// parseInt safely converts a string to an integer (returns -1 if empty or invalid)
func parseInt(value string) int {
	if value == "" {
		return -1
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return -1
	}
	return intValue
}
