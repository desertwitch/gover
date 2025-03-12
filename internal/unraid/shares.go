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
}

func (s *Share) GetName() string {
	return s.Name
}

func (s *Share) GetUseCache() string {
	return s.UseCache
}

func (s *Share) GetCachePool() *Pool {
	return s.CachePool
}

func (s *Share) GetCachePool2() *Pool {
	return s.CachePool
}

func (s *Share) GetAllocator() string {
	return s.Allocator
}

func (s *Share) GetSplitLevel() int {
	return s.SplitLevel
}

func (s *Share) GetSpaceFloor() uint64 {
	return s.SpaceFloor
}

func (s *Share) GetDisableCOW() bool {
	return s.DisableCOW
}

func (s *Share) GetIncludedDisks() map[string]*Disk {
	return s.IncludedDisks
}

func (s *Share) GetExcludedDisks() map[string]*Disk {
	return s.ExcludedDisks
}

func (u *Handler) establishShares(disks map[string]*Disk, pools map[string]*Pool) (map[string]*Share, error) {
	basePath := ConfigDirShares

	if exists, err := u.FSHandler.Exists(basePath); !exists {
		return nil, fmt.Errorf("(unraid-shares) config dir does not exist (%s): %w", basePath, err)
	}

	shares := make(map[string]*Share)

	files, err := u.FSHandler.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("(unraid-shares) failed to readdir: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".cfg") {
			filePath := filepath.Join(basePath, file.Name())
			nameWithoutExt := strings.TrimSuffix(file.Name(), ".cfg")

			configMap, err := u.ConfigHandler.ReadGeneric(filePath)
			if err != nil {
				return nil, fmt.Errorf("(unraid-shares) failed to read config (%s): %w", filePath, err)
			}

			share := &Share{
				Name:       nameWithoutExt,
				UseCache:   u.ConfigHandler.MapKeyToString(configMap, SettingShareUseCache),
				Allocator:  u.ConfigHandler.MapKeyToString(configMap, SettingShareAllocator),
				DisableCOW: strings.ToLower(u.ConfigHandler.MapKeyToString(configMap, SettingShareCOW)) == "no",
				SplitLevel: u.ConfigHandler.MapKeyToInt(configMap, SettingShareSplitLevel),
				SpaceFloor: u.ConfigHandler.MapKeyToUInt64(configMap, SettingShareFloor),
			}

			cachepool, err := findPool(pools, u.ConfigHandler.MapKeyToString(configMap, SettingShareCachePool))
			if err != nil {
				return nil, fmt.Errorf("(unraid-shares) failed to deref primary cache for share (%s): %w", nameWithoutExt, err)
			}
			share.CachePool = cachepool

			cachepool2, err := findPool(pools, u.ConfigHandler.MapKeyToString(configMap, SettingShareCachePool2))
			if err != nil {
				return nil, fmt.Errorf("(unraid-shares) failed to deref secondary cache for share (%s): %w", nameWithoutExt, err)
			}
			share.CachePool2 = cachepool2

			includedDisks, err := findDisks(disks, u.ConfigHandler.MapKeyToString(configMap, SettingShareIncludeDisks))
			if err != nil {
				return nil, fmt.Errorf("(unraid-shares) failed to deref included disks for share (%s): %w", nameWithoutExt, err)
			}
			if includedDisks != nil {
				share.IncludedDisks = includedDisks
			} else {
				// If nil, assume all disks are included
				share.IncludedDisks = disks
			}

			excludedDisks, err := findDisks(disks, u.ConfigHandler.MapKeyToString(configMap, SettingShareExcludeDisks))
			if err != nil {
				return nil, fmt.Errorf("(unraid-shares) failed to deref excluded disks for share (%s): %w", nameWithoutExt, err)
			}
			share.ExcludedDisks = excludedDisks

			shares[share.Name] = share
		}
	}

	return shares, nil
}
