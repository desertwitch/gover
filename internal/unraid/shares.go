package unraid

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Share struct {
	Name          string
	UseCache      string
	CachePool     *Pool
	CachePool2    *Pool
	Allocator     string
	SplitLevel    int
	SpaceFloor    uint64
	DisableCOW    bool
	IncludedDisks map[string]*Disk
	ExcludedDisks map[string]*Disk
	CFGFile       string
}

// TO-DO: Refactor into establishShare() and establishShares().
func (u *Handler) EstablishShares(disks map[string]*Disk, pools map[string]*Pool) (map[string]*Share, error) {
	basePath := ConfigDirShares

	if exists, err := u.FSOps.Exists(basePath); !exists {
		return nil, fmt.Errorf("share config dir does not exist: %w", err)
	}

	shares := make(map[string]*Share)

	files, err := u.FSOps.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read share config dir: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".cfg") {
			filePath := filepath.Join(basePath, file.Name())
			nameWithoutExt := strings.TrimSuffix(file.Name(), ".cfg")

			configMap, err := u.ConfigOps.ReadGeneric(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read share config %s: %w", filePath, err)
			}

			share := &Share{
				Name:       nameWithoutExt,
				CFGFile:    filePath,
				UseCache:   u.ConfigOps.MapKeyToString(configMap, SettingShareUseCache),
				Allocator:  u.ConfigOps.MapKeyToString(configMap, SettingShareAllocator),
				DisableCOW: strings.ToLower(u.ConfigOps.MapKeyToString(configMap, SettingShareCOW)) == "no",
				SplitLevel: u.ConfigOps.MapKeyToInt(configMap, SettingShareSplitLevel),
				SpaceFloor: u.ConfigOps.MapKeyToUInt64(configMap, SettingShareFloor),
			}

			cachepool, err := findPool(pools, u.ConfigOps.MapKeyToString(configMap, SettingShareCachePool))
			if err != nil {
				return nil, fmt.Errorf("failed to dereference primary cache for share %s: %w", nameWithoutExt, err)
			}
			share.CachePool = cachepool

			cachepool2, err := findPool(pools, u.ConfigOps.MapKeyToString(configMap, SettingShareCachePool2))
			if err != nil {
				return nil, fmt.Errorf("failed to dereference secondary cache for share %s: %w", nameWithoutExt, err)
			}
			share.CachePool2 = cachepool2

			includedDisks, err := findDisks(disks, u.ConfigOps.MapKeyToString(configMap, SettingShareIncludeDisks))
			if err != nil {
				return nil, fmt.Errorf("failed to dereference included disks for share %s: %w", nameWithoutExt, err)
			}
			if includedDisks != nil {
				share.IncludedDisks = includedDisks
			} else {
				// If nil, assume all disks are included
				share.IncludedDisks = disks
			}

			excludedDisks, err := findDisks(disks, u.ConfigOps.MapKeyToString(configMap, SettingShareExcludeDisks))
			if err != nil {
				return nil, fmt.Errorf("failed to dereference excluded disks for share %s: %w", nameWithoutExt, err)
			}
			share.ExcludedDisks = excludedDisks

			shares[share.Name] = share
		}
	}

	return shares, nil
}
