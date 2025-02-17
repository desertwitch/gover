package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func proposeArrayDestination(m *Moveable) (*UnraidDisk, error) {
	includedDisks := m.Share.IncludedDisks
	excludedDisks := m.Share.ExcludedDisks

	if m.Share.SplitLevel >= 0 {
		returnDisks, err := allocateDisksBySplitLevel(m)
		if err != nil {
			return nil, fmt.Errorf("failed allocating by split level: %w", err)
		}
		if returnDisks != nil {
			includedDisks = returnDisks
		}
	}

	switch allocationMethod := m.Share.Allocator; allocationMethod {
	case "highwater":
		ret, err := allocateHighWaterDisk(m, includedDisks, excludedDisks)
		if err != nil {
			return nil, fmt.Errorf("failed allocating by high water: %w", err)
		}
		return ret, nil

	case "fillup":
		ret, err := allocateFillUpDisk(m, includedDisks, excludedDisks)
		if err != nil {
			return nil, fmt.Errorf("failed allocating by fillup: %w", err)
		}
		return ret, nil

	case "mostfree":
		ret, err := allocateMostFreeDisk(m, includedDisks, excludedDisks)
		if err != nil {
			return nil, fmt.Errorf("failed allocating by mostfree: %w", err)
		}
		return ret, nil

	default:
		return nil, fmt.Errorf("no allocation method given in configuration")
	}
}

func allocateMostFreeDisk(m *Moveable, includedDisks map[string]*UnraidDisk, excludedDisks map[string]*UnraidDisk) (*UnraidDisk, error) {
	diskStats := make(map[*UnraidDisk]DiskStats)
	var disks []*UnraidDisk

	for name, disk := range includedDisks {
		if _, exists := excludedDisks[name]; exists {
			continue
		}

		stats, err := getDiskUsage(disk.FSPath)
		if err != nil {
			continue
		}
		diskStats[disk] = stats

		disks = append(disks, disk)
	}

	sort.Slice(disks, func(i, j int) bool {
		return diskStats[disks[i]].FreeSpace > diskStats[disks[j]].FreeSpace
	})

	for _, disk := range disks {
		if enoughSpace, _ := hasEnoughFreeSpace(disk, m.Share.SpaceFloor, m.Metadata.Size); enoughSpace {
			return disk, nil
		}
	}

	return nil, nil
}

func allocateFillUpDisk(m *Moveable, includedDisks map[string]*UnraidDisk, excludedDisks map[string]*UnraidDisk) (*UnraidDisk, error) {
	diskStats := make(map[*UnraidDisk]DiskStats)
	var disks []*UnraidDisk

	for name, disk := range includedDisks {
		if _, exists := excludedDisks[name]; exists {
			continue
		}

		stats, err := getDiskUsage(disk.FSPath)
		if err != nil {
			continue
		}
		diskStats[disk] = stats

		disks = append(disks, disk)
	}

	sort.Slice(disks, func(i, j int) bool {
		return diskStats[disks[i]].FreeSpace < diskStats[disks[j]].FreeSpace
	})

	for _, disk := range disks {
		enoughSpace, _ := hasEnoughFreeSpace(disk, m.Share.SpaceFloor, m.Metadata.Size)
		if enoughSpace && diskStats[disk].FreeSpace > m.Share.SpaceFloor {
			return disk, nil
		}
	}

	return nil, nil
}

func allocateHighWaterDisk(m *Moveable, includedDisks map[string]*UnraidDisk, excludedDisks map[string]*UnraidDisk) (*UnraidDisk, error) {
	diskStats := make(map[*UnraidDisk]DiskStats)
	var disks []*UnraidDisk

	var maxDiskSize int64

	for name, disk := range includedDisks {
		if _, exists := excludedDisks[name]; exists {
			continue
		}

		stats, err := getDiskUsage(disk.FSPath)
		if err != nil {
			continue
		}
		diskStats[disk] = stats

		if stats.TotalSize > maxDiskSize {
			maxDiskSize = stats.TotalSize
		}

		disks = append(disks, disk)
	}

	if maxDiskSize == 0 {
		return nil, fmt.Errorf("failed to retrieve disk space information for any disk")
	}

	highWaterMark := maxDiskSize / 2

	for highWaterMark > 0 {
		sort.Slice(disks, func(i, j int) bool {
			return diskStats[disks[i]].FreeSpace < diskStats[disks[j]].FreeSpace
		})
		for _, disk := range disks {
			enoughSpace, _ := hasEnoughFreeSpace(disk, m.Share.SpaceFloor, m.Metadata.Size)
			if stats, exists := diskStats[disk]; exists && enoughSpace && stats.FreeSpace >= highWaterMark {
				return disk, nil
			}
		}
		highWaterMark /= 2
	}

	return nil, nil
}

func allocateDisksBySplitLevel(m *Moveable) (map[string]*UnraidDisk, error) {
	matches := make(map[int]map[string]*UnraidDisk)

	mainMatches, mainLevel, err := findDisksBySplitLevel(m)
	if err != nil {
		return nil, fmt.Errorf("failed allocating disk by split level: %w", err)
	}

	if len(mainMatches) > 0 {
		matches[mainLevel] = make(map[string]*UnraidDisk)
		for _, disk := range mainMatches {
			matches[mainLevel][disk.Name] = disk
		}
	}

	if len(m.Hardlinks) > 0 {
		for _, s := range m.Hardlinks {
			subMatches, subLevel, err := findDisksBySplitLevel(s)
			if err != nil {
				return nil, fmt.Errorf("failed suballocating disk by split level: %w", err)
			}
			if len(subMatches) > 0 {
				if matches[subLevel] == nil {
					matches[subLevel] = make(map[string]*UnraidDisk)
				}
				for _, disk := range subMatches {
					matches[subLevel][disk.Name] = disk
				}
			}
		}

		maxKey := -1
		for key := range matches {
			if key > maxKey {
				maxKey = key
			}
		}

		if bestMatch, exists := matches[maxKey]; exists {
			return bestMatch, nil
		}

		return nil, nil
	}

	if len(matches[mainLevel]) > 0 {
		return matches[mainLevel], nil
	}

	return nil, nil
}

func findDisksBySplitLevel(m *Moveable) ([]*UnraidDisk, int, error) {
	foundDisks := []*UnraidDisk{}
	path := filepath.Dir(m.Path)

	relPath, err := filepath.Rel(m.Source.GetFSPath(), path)
	if err != nil {
		return nil, -1, fmt.Errorf("failed deriving subpath: %w", err)
	}

	pathParts := strings.Split(relPath, string(os.PathSeparator))
	splitLevel := len(pathParts)

	if splitLevel == 0 {
		return nil, -1, fmt.Errorf("calculated split level of zero: %s", path)
	}

	maxLevel := m.Share.SplitLevel

	if splitLevel <= maxLevel {
		return nil, -1, nil
	} else {
		for i := len(pathParts[maxLevel:]); i > 0; i-- {
			subPath := filepath.Join(pathParts[:maxLevel+i]...)
			found := false
			for name, disk := range m.Share.IncludedDisks {
				if _, exists := m.Share.ExcludedDisks[name]; exists {
					continue
				}
				dirToCheck := filepath.Join(disk.FSPath, subPath)
				if _, err := os.Stat(dirToCheck); err == nil {
					if enoughSpace, _ := hasEnoughFreeSpace(disk, m.Share.SpaceFloor, m.Metadata.Size); enoughSpace {
						foundDisks = append(foundDisks, disk)
						found = true
					}
				}
			}
			if found {
				return foundDisks, i, nil
			}
		}
		return nil, -1, nil
	}
}
