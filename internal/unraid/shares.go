package unraid

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type UnraidShare struct {
	Name          string
	UseCache      string
	CachePool     *UnraidPool
	CachePool2    *UnraidPool
	Allocator     string
	SplitLevel    int
	SpaceFloor    int64
	DisableCOW    bool
	IncludedDisks map[string]*UnraidDisk
	ExcludedDisks map[string]*UnraidDisk
	CFGFile       string
}

// establishShares returns a map of pointers to established Unraid shares
// TO-DO: Refactor into establishShare() and establishShares()
func establishShares(disks map[string]*UnraidDisk, pools map[string]*UnraidPool, osa osAdapter) (map[string]*UnraidShare, error) {
	basePath := ConfigDirShares

	if _, err := osa.Stat(basePath); errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("share config dir does not exist: %w", err)
	}

	shares := make(map[string]*UnraidShare)

	files, err := osa.ReadDir(basePath)
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
