package unraid

import (
	"fmt"
)

const (
	ArrayStateFile = "/var/local/emhttp/var.ini"

	ConfigDirShares = "/boot/config/shares"
	ConfigDirPools  = "/boot/config/pools"

	BasePathMounts = "/mnt/"
	PatternDisks   = `^disk[1-9][0-9]?$`

	AllocHighWater = "highwater"
	AllocMostFree  = "mostfree"
	AllocFillUp    = "fillup"

	SettingShareUseCache   = "shareUseCache"
	SettingShareAllocator  = "shareAllocator"
	SettingShareCOW        = "shareCOW"
	SettingShareSplitLevel = "shareSplitLevel"
	SettingShareFloor      = "shareFloor"

	SettingShareCachePool    = "shareCachePool"
	SettingShareCachePool2   = "shareCachePool2"
	SettingShareIncludeDisks = "shareInclude"
	SettingShareExcludeDisks = "shareExclude"

	StateArrayStatus    = "mdState"
	StateTurboSetting   = "md_write_method"
	StateParityPosition = "mdResyncPos"
)

type UnraidSystem struct {
	Array  *UnraidArray
	Pools  map[string]*UnraidPool
	Shares map[string]*UnraidShare
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
