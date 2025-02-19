package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
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
	return nil, fmt.Errorf("configured pool %s does not exist", poolName)
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
			return nil, fmt.Errorf("configured disk %s does not exist", name)
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

// parseInt64 safely converts a string to a 64-bit integer (returns -1 if empty or invalid)
func parseInt64(value string) int64 {
	if value == "" {
		return -1
	}
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return -1
	}
	return intValue
}

// removeInternalLinks removes symbolic and hardlink moveable pointers from a slice of moveable pointers
func removeInternalLinks(moveables []*Moveable) []*Moveable {
	var ms []*Moveable

	for _, m := range moveables {
		if !m.Symlink && !m.Hardlink {
			ms = append(ms, m)
		}
	}

	return ms
}

func existsOnAnyShareDisk(m *Moveable) (bool, []*UnraidDisk) {
	var foundDisks []*UnraidDisk

	for name, disk := range m.Share.IncludedDisks {
		if _, exists := m.Share.ExcludedDisks[name]; exists {
			continue
		}
		relPath, err := filepath.Rel(m.Source.GetFSPath(), m.SourcePath)
		if err != nil {
			slog.Warn("Skipped disk for destination file existence consideration", "err", err, "job", m.SourcePath, "share", m.Share.Name)
			continue
		}
		probePath := filepath.Join(disk.FSPath, relPath)
		if _, err := os.Stat(probePath); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				slog.Warn("Skipped disk for destination file existence consideration", "path", probePath, "err", err, "job", m.SourcePath, "share", m.Share.Name)
			}
			continue
		}
		foundDisks = append(foundDisks, disk)
	}

	return len(foundDisks) > 0, foundDisks
}
