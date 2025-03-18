package unraid

import (
	"os"
)

const (
	ArrayStateFile        = "/var/local/emhttp/var.ini"
	GlobalShareConfigFile = "/boot/config/share.cfg"

	ConfigDirShares = "/boot/config/shares"
	ConfigDirPools  = "/boot/config/pools"

	BasePathMounts = "/mnt/"
	PatternDisks   = `^disk[1-9][0-9]?$`

	SettingGlobalShareIncludes = "shareUserInclude"
	SettingGlobalShareExcludes = "shareUserExclude"

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

type fsProvider interface {
	Exists(path string) (bool, error)
	ReadDir(name string) ([]os.DirEntry, error)
}

type configProvider interface {
	ReadGeneric(filenames ...string) (envMap map[string]string, err error)
	MapKeyToString(envMap map[string]string, key string) string
	MapKeyToInt(envMap map[string]string, key string) int
	MapKeyToInt64(envMap map[string]string, key string) int64
	MapKeyToUInt64(envMap map[string]string, key string) uint64
}

type Handler struct {
	fsHandler     fsProvider
	configHandler configProvider
}

func NewHandler(fsHandler fsProvider, configHandler configProvider) *Handler {
	return &Handler{
		fsHandler:     fsHandler,
		configHandler: configHandler,
	}
}
