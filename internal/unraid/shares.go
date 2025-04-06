package unraid

import (
	"fmt"
	"maps"
	"path/filepath"
	"strings"
)

// Share is an Unraid share, as part of an Unraid [System].
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

// GetName returns the share's name.
func (s *Share) GetName() string {
	return s.Name
}

// GetUseCache returns the caching setting for the share.
func (s *Share) GetUseCache() string {
	return s.UseCache
}

// GetCachePool returns the primary cache [Pool]. If it is nil, the primary
// cache pool is the [Array] (just for interpretation, the pointer is nil).
func (s *Share) GetCachePool() *Pool {
	return s.CachePool
}

// GetCachePool2 returns the secondary cache [Pool]. If it is nil, the secondary
// cache pool is the [Array] (just for interpretation, the pointer is nil).
func (s *Share) GetCachePool2() *Pool {
	return s.CachePool2
}

// GetAllocator returns the allocation method for the share.
func (s *Share) GetAllocator() string {
	return s.Allocator
}

// GetSplitLevel returns the split level setting for the share.
func (s *Share) GetSplitLevel() int {
	return s.SplitLevel
}

// GetSpaceFloor returns the minimum free space setting for the share.
func (s *Share) GetSpaceFloor() uint64 {
	return s.SpaceFloor
}

// GetDisableCOW returns if CoW should be disabled for the share.
func (s *Share) GetDisableCOW() bool {
	return s.DisableCOW
}

// GetIncludedDisks returns a copy of the internal map holding pointers to all
// included [Disk].
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

// includesExcludesConfig holds information about [Disk] inclusions and
// exclusions, both for an individual [Share] and for the whole [System].
type includesExcludesConfig struct {
	shareIncludes  map[string]*Disk
	shareExcludes  map[string]*Disk
	globalIncludes map[string]*Disk
	globalExcludes map[string]*Disk
}

// establishShares returns a map (map[shareName]*Share) to all Unraid [Share].
// It is the principal method for reading all share information from the system.
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

			cachepool, err := findPool(u.configHandler.MapKeyToString(configMap, SettingShareCachePool), pools)
			if err != nil {
				return nil, fmt.Errorf("(unraid-shares) failed to deref primary cache for share (%s): %w", nameWithoutExt, err)
			}
			share.CachePool = cachepool

			cachepool2, err := findPool(u.configHandler.MapKeyToString(configMap, SettingShareCachePool2), pools)
			if err != nil {
				return nil, fmt.Errorf("(unraid-shares) failed to deref secondary cache for share (%s): %w", nameWithoutExt, err)
			}
			share.CachePool2 = cachepool2

			shareIncludes, err := findDisks(u.configHandler.MapKeyToString(configMap, SettingShareIncludeDisks), disks)
			if err != nil {
				return nil, fmt.Errorf("(unraid-shares) failed to deref included disks for share (%s): %w", nameWithoutExt, err)
			}

			shareExcludes, err := findDisks(u.configHandler.MapKeyToString(configMap, SettingShareExcludeDisks), disks)
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

// establishGlobalIncludesExcludes returns a pointer to a new
// [includesExcludesConfig], with the global inclusions and exclusions already
// populated for later consideration with individual [Share] inclusions and
// exclusions.
func (u *Handler) establishGlobalIncludesExcludes(disks map[string]*Disk) (*includesExcludesConfig, error) {
	configMap, err := u.configHandler.ReadGeneric(GlobalShareConfigFile)
	if err != nil {
		return nil, fmt.Errorf("(unraid-shares) failed to read global share config (%s): %w", GlobalShareConfigFile, err)
	}

	globalIncludes, err := findDisks(u.configHandler.MapKeyToString(configMap, SettingGlobalShareIncludes), disks)
	if err != nil {
		return nil, fmt.Errorf("(unraid-shares) failed to deref global included disks: %w", err)
	}

	globalExcludes, err := findDisks(u.configHandler.MapKeyToString(configMap, SettingGlobalShareExcludes), disks)
	if err != nil {
		return nil, fmt.Errorf("(unraid-shares) failed to deref global excluded disks: %w", err)
	}

	return &includesExcludesConfig{
		globalIncludes: globalIncludes,
		globalExcludes: globalExcludes,
	}, nil
}

// establishShareIncludes considers a [includesExcludesConfig] and returns a map
// (map[diskName]*Disk) of only the effectively included [Disk] for a [Share].
func (u *Handler) establishShareIncludes(allDisks map[string]*Disk, config *includesExcludesConfig) map[string]*Disk {
	shareIncludes := make(map[string]*Disk)
	globalIncludes := make(map[string]*Disk)

	if config.shareIncludes == nil {
		maps.Copy(shareIncludes, allDisks)
	} else {
		maps.Copy(shareIncludes, config.shareIncludes)
	}

	if config.globalIncludes == nil {
		maps.Copy(globalIncludes, allDisks)
	} else {
		maps.Copy(globalIncludes, config.globalIncludes)
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
