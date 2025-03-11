package storage

type Storage interface {
	GetName() string
	GetFSPath() string
}

type Disk interface {
	IsDisk() bool
	GetName() string
	GetFSPath() string
}

type Pool interface {
	IsPool() bool
	GetName() string
	GetFSPath() string
}

type Share interface {
	GetName() string
	GetUseCache() string
	GetCachePool() Pool
	GetCachePool2() Pool
	GetAllocator() string
	GetSplitLevel() int
	GetSpaceFloor() uint64
	GetDisableCOW() bool
	GetIncludedDisks() map[string]Disk
	GetExcludedDisks() map[string]Disk
}
