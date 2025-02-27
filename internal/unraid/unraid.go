package unraid

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

// establishDisks returns a map of pointers to established Unraid disks
func establishDisks() (map[string]*UnraidDisk, error) {
	basePath := BasePathMounts
	diskPattern := regexp.MustCompile(PatternDisks)

	disks := make(map[string]*UnraidDisk)

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mounts at %s: %w", basePath, err)
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
// TO-DO: Refactor into establishPool() and establishPools()
func establishPools() (map[string]*UnraidPool, error) {
	basePath := ConfigDirPools

	if _, err := os.Stat(basePath); errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("pool config dir does not exist: %w", err)
	}

	pools := make(map[string]*UnraidPool)

	files, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pool config dir: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".cfg") {
			cfgPath := filepath.Join(basePath, file.Name())
			nameWithoutExt := strings.TrimSuffix(file.Name(), ".cfg")

			fsPath := filepath.Join("/mnt", nameWithoutExt)
			if _, err := os.Stat(fsPath); errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("pool mount %s does not exist: %w", fsPath, err)
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
// TO-DO: Refactor into establishShare() and establishShares()
func establishShares(disks map[string]*UnraidDisk, pools map[string]*UnraidPool) (map[string]*UnraidShare, error) {
	basePath := ConfigDirShares

	if _, err := os.Stat(basePath); errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("share config dir does not exist: %w", err)
	}

	shares := make(map[string]*UnraidShare)

	files, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read share config dir: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".cfg") {
			filePath := filepath.Join(basePath, file.Name())
			nameWithoutExt := strings.TrimSuffix(file.Name(), ".cfg")

			configMap, err := godotenv.Read(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read share config %s: %w", filePath, err)
			}

			share := &UnraidShare{
				Name:       nameWithoutExt,
				CFGFile:    filePath,
				UseCache:   getConfigValue(configMap, SettingShareUseCache),
				Allocator:  getConfigValue(configMap, SettingShareAllocator),
				DisableCOW: strings.ToLower(getConfigValue(configMap, SettingShareCOW)) == "no",
				SplitLevel: parseInt(getConfigValue(configMap, SettingShareSplitLevel)),
				SpaceFloor: parseInt64(getConfigValue(configMap, SettingShareFloor)),
			}

			cachepool, err := findPool(pools, getConfigValue(configMap, SettingShareCachePool))
			if err != nil {
				return nil, fmt.Errorf("failed to dereference primary cache for share %s: %w", nameWithoutExt, err)
			}
			share.CachePool = cachepool

			cachepool2, err := findPool(pools, getConfigValue(configMap, SettingShareCachePool2))
			if err != nil {
				return nil, fmt.Errorf("failed to dereference secondary cache for share %s: %w", nameWithoutExt, err)
			}
			share.CachePool2 = cachepool2

			includedDisks, err := findDisks(disks, getConfigValue(configMap, SettingShareIncludeDisks))
			if err != nil {
				return nil, fmt.Errorf("failed to dereference included disks for share %s: %w", nameWithoutExt, err)
			}
			if includedDisks != nil {
				share.IncludedDisks = includedDisks
			} else {
				// If nil, assume all disks are included
				share.IncludedDisks = disks
			}

			excludedDisks, err := findDisks(disks, getConfigValue(configMap, SettingShareExcludeDisks))
			if err != nil {
				return nil, fmt.Errorf("failed to dereference excluded disks for share %s: %w", nameWithoutExt, err)
			}
			share.ExcludedDisks = excludedDisks

			shares[share.Name] = share
		}
	}

	return shares, nil
}

// establishArray returns a pointer to an established Unraid array
func establishArray(disks map[string]*UnraidDisk) (*UnraidArray, error) {
	stateFile := ArrayStateFile

	configMap, err := godotenv.Read(stateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load disk state file %s: %w", stateFile, err)
	}

	array := &UnraidArray{
		Disks:         disks,
		Status:        getConfigValue(configMap, StateArrayStatus),
		TurboSetting:  getConfigValue(configMap, StateTurboSetting),
		ParityRunning: parseInt(getConfigValue(configMap, StateParityPosition)) > 0,
	}

	return array, nil
}

// establishSystem returns a pointer to an established Unraid system
func EstablishSystem() (*UnraidSystem, error) {
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
