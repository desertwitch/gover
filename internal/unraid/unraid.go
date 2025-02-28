package unraid

import (
	"fmt"
	"os"
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

type osProvider interface {
	ReadDir(name string) ([]os.DirEntry, error)
	Stat(name string) (os.FileInfo, error)
}

type configProvider interface {
	ReadGeneric(filenames ...string) (envMap map[string]string, err error)
	MapKeyToString(envMap map[string]string, key string) string
	MapKeyToInt(envMap map[string]string, key string) int
	MapKeyToInt64(envMap map[string]string, key string) int64
}

type UnraidStoreable interface {
	GetName() string
	GetFSPath() string
	IsActiveTransfer() bool
	SetActiveTransfer(bool)
}

type UnraidSystem struct {
	Array  *UnraidArray
	Pools  map[string]*UnraidPool
	Shares map[string]*UnraidShare
}

type UnraidHandler struct {
	OSOps     osProvider
	ConfigOps configProvider
}

// establishSystem returns a pointer to an established Unraid system
func (u *UnraidHandler) EstablishSystem() (*UnraidSystem, error) {
	disks, err := u.EstablishDisks()
	if err != nil {
		return nil, fmt.Errorf("failed establishing disks: %w", err)
	}

	pools, err := u.EstablishPools()
	if err != nil {
		return nil, fmt.Errorf("failed establishing pools: %w", err)
	}

	shares, err := u.EstablishShares(disks, pools)
	if err != nil {
		return nil, fmt.Errorf("failed establishing shares: %w", err)
	}

	array, err := u.EstablishArray(disks)
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
