package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// getConfigValue returns an string element of a string map or "" if not existing
func getConfigValue(envMap map[string]string, key string) string {
	if value, exists := envMap[key]; exists {
		return value
	}
	return ""
}

// findPool dereferences a textual pool name into a pool pointer
func findPool(pools map[string]*UnraidPool, poolName string) (*UnraidPool, error) {
	if poolName == "" {
		return nil, nil
	}
	if pool, exists := pools[poolName]; exists {
		return pool, nil
	}
	return nil, fmt.Errorf("configured pool %s not found in mounted pools", poolName)
}

// findDisks dereferences a list of textual disk names into a map of disk pointers
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

func parseStateFile(filename string) (map[string]map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := make(map[string]map[string]string)

	currentSection := "global"
	data[currentSection] = make(map[string]string)

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, `["`) && strings.HasSuffix(line, `"]`) {
			currentSection = line[2 : len(line)-2]
			if _, exists := data[currentSection]; !exists {
				data[currentSection] = make(map[string]string)
			}
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Warning: Invalid format on line %d: %s\n", lineNumber, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
			value = value[1 : len(value)-1]
		}

		data[currentSection][key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return data, nil
}
