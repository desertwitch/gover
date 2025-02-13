package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

// establishDisks returns a map of pointers to established Unraid disks
func establishDisks() (map[string]*UnraidDisk, error) {
	basePath := "/mnt/"
	diskPattern := regexp.MustCompile(`^disk[1-9][0-9]?$`)

	disks := make(map[string]*UnraidDisk)

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mount directory %s: %w", basePath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() && diskPattern.MatchString(entry.Name()) {
			disk := &UnraidDisk{
				Name:           entry.Name(),
				FSPath:         filepath.Join(basePath, entry.Name()),
				ActiveTransfer: false,
			}
			disks[disk.Name] = disk
		}
	}

	return disks, nil
}

// establishPools returns a map of pointers to established Unraid pools
func establishPools() (map[string]*UnraidPool, error) {
	basePath := "/boot/config/pools"

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("pool configuration directory %s does not exist: %w", basePath, err)
	}

	pools := make(map[string]*UnraidPool)

	files, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pool configuration directory %s: %w", basePath, err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".cfg") {
			cfgPath := filepath.Join(basePath, file.Name())
			nameWithoutExt := strings.TrimSuffix(file.Name(), ".cfg")

			fsPath := filepath.Join("/mnt", nameWithoutExt)
			if _, err := os.Lstat(fsPath); os.IsNotExist(err) {
				return nil, fmt.Errorf("pool mount %s does not exist despite configuration at %s: %w", fsPath, cfgPath, err)
			}

			pool := &UnraidPool{
				Name:           nameWithoutExt,
				FSPath:         fsPath,
				CFGFile:        cfgPath,
				ActiveTransfer: false,
			}

			pools[pool.Name] = pool
		}
	}

	return pools, nil
}

// establishShares returns a map of pointers to established Unraid shares
func establishShares(disks map[string]*UnraidDisk, pools map[string]*UnraidPool) (map[string]*UnraidShare, error) {
	basePath := "/boot/config/shares"

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("share configuration directory %s does not exist: %w", basePath, err)
	}

	shares := make(map[string]*UnraidShare)

	files, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read share configuration directory %s: %w", basePath, err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".cfg") {
			filePath := filepath.Join(basePath, file.Name())
			nameWithoutExt := strings.TrimSuffix(file.Name(), ".cfg")

			configMap, err := godotenv.Read(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read share configuration %s: %w", filePath, err)
			}

			share := &UnraidShare{
				Name:       nameWithoutExt,
				CFGFile:    filePath,
				UseCache:   getConfigValue(configMap, "shareUseCache"),
				Allocator:  getConfigValue(configMap, "shareAllocator"),
				DisableCOW: strings.ToLower(getConfigValue(configMap, "shareCOW")) == "no",
				SplitLevel: parseInt(getConfigValue(configMap, "shareSplitLevel")),
				SpaceFloor: parseInt(getConfigValue(configMap, "shareFloor")),
			}

			cachepool, err := findPool(pools, getConfigValue(configMap, "shareCachePool"))
			if err != nil {
				return nil, fmt.Errorf("failed to dereference primary cache pool for share %s: %w", nameWithoutExt, err)
			}
			share.CachePool = cachepool

			cachepool2, err := findPool(pools, getConfigValue(configMap, "shareCachePool2"))
			if err != nil {
				return nil, fmt.Errorf("failed to dereference secondary cache pool for share %s: %w", nameWithoutExt, err)
			}
			share.CachePool2 = cachepool2

			includedDisks, err := findDisks(disks, getConfigValue(configMap, "shareInclude"))
			if err != nil {
				return nil, fmt.Errorf("failed to dereference included disks for share %s: %w", nameWithoutExt, err)
			}
			if includedDisks != nil {
				share.IncludedDisks = includedDisks
			} else {
				// If nil, assume all disks are included
				share.IncludedDisks = disks
			}

			excludedDisks, err := findDisks(disks, getConfigValue(configMap, "shareExclude"))
			if err != nil {
				return nil, fmt.Errorf("failed to dereference excluded disks for share %s: %w", nameWithoutExt, err)
			}
			if excludedDisks != nil {
				share.ExcludedDisks = excludedDisks
			} else {
				// If nil, assume no disks are excluded
				share.ExcludedDisks = make(map[string]*UnraidDisk)
			}

			shares[share.Name] = share
		}
	}

	return shares, nil
}

// establishArray returns a pointer to an established Unraid array
func establishArray(disks map[string]*UnraidDisk) (*UnraidArray, error) {
	stateFile := "/var/local/emhttp/var.ini"

	configMap, err := godotenv.Read(stateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load disk state file %s: %w", stateFile, err)
	}

	array := &UnraidArray{
		Disks:         disks,
		Status:        getConfigValue(configMap, "mdState"),
		TurboSetting:  getConfigValue(configMap, "md_write_method"),
		ParityRunning: parseInt(getConfigValue(configMap, "mdResyncPos")) > 0,
	}

	return array, nil
}

// establishSystem returns a pointer to an established Unraid system
func establishSystem() (*UnraidSystem, error) {
	disks, err := establishDisks()
	if err != nil {
		return nil, fmt.Errorf("failed establishing disks: %w", err)
	}

	pools, err := establishPools()
	if err != nil {
		return nil, fmt.Errorf("failed establishing pools: %w", err)
	}

	shares, err := establishShares(disks, pools)
	if err != nil {
		return nil, fmt.Errorf("failed establishing shares: %w", err)
	}

	array, err := establishArray(disks)
	if err != nil {
		return nil, fmt.Errorf("failed establishing array: %w", err)
	}

	system := &UnraidSystem{
		Array:  array,
		Pools:  pools,
		Shares: shares,
	}

	return system, nil
}
