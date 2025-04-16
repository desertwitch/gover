package schema

// Storage describes methods a [Storage] type needs to have.
type Storage interface {
	GetName() string
	GetFSPath() string
}

// Disk describes methods that a [Storage] of type [Disk] needs to have.
type Disk interface {
	IsDisk() bool
	GetName() string
	GetFSPath() string
}

// Pool describes methods that a [Storage] of type [Pool] needs to have.
type Pool interface {
	IsPool() bool
	GetName() string
	GetFSPath() string
}

// Share describes methods that a [Share] needs to have.
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
	GetPipeline() Pipeline
}
