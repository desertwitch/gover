package unraid

const (
	// ArrayStateFile contains state information about the [Array].
	ArrayStateFile = "/var/local/emhttp/var.ini"

	// GlobalShareConfigFile contains the global [Share] settings.
	GlobalShareConfigFile = "/boot/config/share.cfg"

	// ConfigDirShares contains all [Share] configurations.
	ConfigDirShares = "/boot/config/shares"

	// ConfigDirPools contains all [Pool] configurations.
	ConfigDirPools = "/boot/config/pools"

	// BasePathMounts is the base path for mountpoints.
	BasePathMounts = "/mnt/"

	// PatternDisks is a regex pattern used when matching for Unraid [Disk].
	PatternDisks = `^disk[1-9][0-9]?$`

	// SettingGlobalShareIncludes is the [GlobalShareConfigFile] configuration
	// key for global [Share] includes.
	SettingGlobalShareIncludes = "shareUserInclude"

	// SettingGlobalShareExcludes is the [GlobalShareConfigFile] configuration
	// key for global [Share] excludes.
	SettingGlobalShareExcludes = "shareUserExclude"

	// SettingShareUseCache is the per-[Share] configuration key for the caching
	// setting.
	SettingShareUseCache = "shareUseCache"

	// SettingShareAllocator is the per-[Share] configuration key for the
	// allocation setting.
	SettingShareAllocator = "shareAllocator"

	// SettingShareCOW is the per-[Share] configuration key for the no-CoW
	// setting.
	SettingShareCOW = "shareCOW"

	// SettingShareSplitLevel is the per-[Share] configuration key for the
	// split-level setting.
	SettingShareSplitLevel = "shareSplitLevel"

	// SettingShareFloor is the per-[Share] configuration key for the minimum
	// free space setting.
	SettingShareFloor = "shareFloor"

	// SettingShareCachePool is the per-[Share] configuration key for the
	// primary (cache) storage.
	SettingShareCachePool = "shareCachePool"

	// SettingShareCachePool2 is the per-[Share] configuration key for the
	// secondary (cache) storage.
	SettingShareCachePool2 = "shareCachePool2"

	// SettingShareIncludeDisks is the per-[Share] configuration key for the
	// share includes.
	SettingShareIncludeDisks = "shareInclude"

	// SettingShareExcludeDisks is the per-[Share] configuration key for the
	// share excludes.
	SettingShareExcludeDisks = "shareExclude"

	// StateArrayStatus is the state information for the [Array] status.
	StateArrayStatus = "mdState"

	// StateTurboSetting is the state information for the turbo setting.
	StateTurboSetting = "md_write_method"

	// StateParityPosition is the state information for the parity operations.
	StateParityPosition = "mdResyncPos"
)
