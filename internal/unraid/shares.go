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
	return s.CachePool2
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

// GetIncludedDisks returns a copy of the internal map holding pointers to all disks.
func (s *Share) GetIncludedDisks() map[string]*Disk {
	if s.IncludedDisks == nil {
		return nil
	}

	disks := make(map[string]*Disk)

	for k, v := range s.IncludedDisks {
		disks[k] = v
	}

	return disks
}

type includesExcludesConfig struct {
	shareIncludes  map[string]*Disk
	shareExcludes  map[string]*Disk
	globalIncludes map[string]*Disk
	globalExcludes map[string]*Disk
}

func (u *Handler) establishShares(disks map[string]*Disk, pools map[string]*Pool) (map[string]*Share, error) {
	basePath := ConfigDirShares

	if exists, err := u.fsHandler.Exists(basePath); !exists {
		return nil, fmt.Errorf("(unraid-shares) config dir does not exist (%s): %w", basePath, err)
	}

	files, err := u.osHandler.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("(unraid-shares) failed to readdir: %w", err)
	}

	includesExcludes, err := u.establishGlobalIncludesExcludes(disks)
	if err != nil {
		return nil, fmt.Errorf("(unraid-shares) failed to establish global share config: %w", err)
	}

	shares := make(map[string]*Share)

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".cfg") {
			filePath := filepath.Join(basePath, file.Name())
			nameWithoutExt := strings.TrimSuffix(file.Name(), ".cfg")

			configMap, err := u.configHandler.ReadGeneric(filePath)
			if err != nil {
				return nil, fmt.Errorf("(unraid-shares) failed to read config (%s): %w", filePath, err)
			}

			share := &Share{
				Name:       nameWithoutExt,
				UseCache:   u.configHandler.MapKeyToString(configMap, SettingShareUseCache),
				Allocator:  u.configHandler.MapKeyToString(configMap, SettingShareAllocator),
				DisableCOW: strings.ToLower(u.configHandler.MapKeyToString(configMap, SettingShareCOW)) == "no",
				SplitLevel: u.configHandler.MapKeyToInt(configMap, SettingShareSplitLevel),
				SpaceFloor: u.configHandler.MapKeyToUInt64(configMap, SettingShareFloor),
			}

			cachepool, err := findPool(pools, u.configHandler.MapKeyToString(configMap, SettingShareCachePool))
			if err != nil {
				return nil, fmt.Errorf("(unraid-shares) failed to deref primary cache for share (%s): %w", nameWithoutExt, err)
			}
			share.CachePool = cachepool

			cachepool2, err := findPool(pools, u.configHandler.MapKeyToString(configMap, SettingShareCachePool2))
			if err != nil {
				return nil, fmt.Errorf("(unraid-shares) failed to deref secondary cache for share (%s): %w", nameWithoutExt, err)
			}
			share.CachePool2 = cachepool2

			shareIncludes, err := findDisks(disks, u.configHandler.MapKeyToString(configMap, SettingShareIncludeDisks))
			if err != nil {
				return nil, fmt.Errorf("(unraid-shares) failed to deref included disks for share (%s): %w", nameWithoutExt, err)
			}

			shareExcludes, err := findDisks(disks, u.configHandler.MapKeyToString(configMap, SettingShareExcludeDisks))
			if err != nil {
				return nil, fmt.Errorf("(unraid-shares) failed to deref excluded disks for share (%s): %w", nameWithoutExt, err)
			}

			includesExcludes.shareIncludes = shareIncludes
			includesExcludes.shareExcludes = shareExcludes

			share.IncludedDisks = u.establishShareIncludes(disks, includesExcludes)

			shares[share.Name] = share
		}
	}

	return shares, nil
}

func (u *Handler) establishGlobalIncludesExcludes(disks map[string]*Disk) (*includesExcludesConfig, error) {
	configMap, err := u.configHandler.ReadGeneric(GlobalShareConfigFile)
	if err != nil {
		return nil, fmt.Errorf("(unraid-shares) failed to read global share config (%s): %w", GlobalShareConfigFile, err)
	}

	globalIncludes, err := findDisks(disks, u.configHandler.MapKeyToString(configMap, SettingGlobalShareIncludes))
	if err != nil {
		return nil, fmt.Errorf("(unraid-shares) failed to deref global included disks: %w", err)
	}

	globalExcludes, err := findDisks(disks, u.configHandler.MapKeyToString(configMap, SettingGlobalShareExcludes))
	if err != nil {
		return nil, fmt.Errorf("(unraid-shares) failed to deref global excluded disks: %w", err)
	}

	return &includesExcludesConfig{
		globalIncludes: globalIncludes,
		globalExcludes: globalExcludes,
	}, nil
}

// establishIncludedDisks returns a map holding pointers to all effectively included disks (excluding any excluded disks).
func (u *Handler) establishShareIncludes(allDisks map[string]*Disk, config *includesExcludesConfig) map[string]*Disk {
	shareIncludes := make(map[string]*Disk)
	globalIncludes := make(map[string]*Disk)

	if config.shareIncludes == nil {
		for k, v := range allDisks {
			shareIncludes[k] = v
		}
	} else {
		for k, v := range config.shareIncludes {
			shareIncludes[k] = v
		}
	}

	if config.globalIncludes == nil {
		for k, v := range allDisks {
			globalIncludes[k] = v
		}
	} else {
		for k, v := range config.globalIncludes {
			globalIncludes[k] = v
		}
	}

	for name := range shareIncludes {
		if _, exists := globalIncludes[name]; !exists {
			delete(shareIncludes, name)
		}
		if _, exists := config.globalExcludes[name]; exists {
			delete(shareIncludes, name)
		}
		if _, exists := config.shareExcludes[name]; exists {
			delete(shareIncludes, name)
		}
	}

	return shareIncludes
}
