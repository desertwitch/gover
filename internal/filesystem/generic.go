package filesystem

type StorageType interface {
	GetName() string
	GetFSPath() string
}

type SystemType interface {
	GetArray() ArrayType
	GetShares() map[string]ShareType
}

type ArrayType interface {
	GetDisks() map[string]DiskType
}
type DiskType interface {
	IsDisk() bool
	GetName() string
	GetFSPath() string
}

type PoolType interface {
	IsPool() bool
	GetName() string
	GetFSPath() string
}

type ShareType interface {
	GetName() string
	GetUseCache() string
	GetCachePool() PoolType
	GetCachePool2() PoolType
	GetAllocator() string
	GetSplitLevel() int
	GetSpaceFloor() uint64
	GetDisableCOW() bool
	GetIncludedDisks() map[string]DiskType
	GetExcludedDisks() map[string]DiskType
}
