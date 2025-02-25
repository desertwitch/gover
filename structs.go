package main

import (
	"golang.org/x/sys/unix"
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

type UnraidArray struct {
	Disks         map[string]*UnraidDisk
	Status        string
	TurboSetting  string
	ParityRunning bool
}

type UnraidStoreable interface {
	GetName() string
	GetFSPath() string
	IsActiveTransfer() bool
	SetActiveTransfer(bool)
}

type UnraidDisk struct {
	Name           string
	FSPath         string
	ActiveTransfer bool
}

func (d *UnraidDisk) GetName() string {
	return d.Name
}

func (d *UnraidDisk) GetFSPath() string {
	return d.FSPath
}

func (d *UnraidDisk) IsActiveTransfer() bool {
	return d.ActiveTransfer
}

func (d *UnraidDisk) SetActiveTransfer(active bool) {
	d.ActiveTransfer = active
}

type UnraidPool struct {
	Name           string
	FSPath         string
	CFGFile        string
	ActiveTransfer bool
}

func (p *UnraidPool) GetName() string {
	return p.Name
}

func (p *UnraidPool) GetFSPath() string {
	return p.FSPath
}

func (p *UnraidPool) IsActiveTransfer() bool {
	return p.ActiveTransfer
}

func (p *UnraidPool) SetActiveTransfer(active bool) {
	p.ActiveTransfer = active
}

type UnraidShare struct {
	Name          string
	UseCache      string
	CachePool     *UnraidPool
	CachePool2    *UnraidPool
	Allocator     string
	SplitLevel    int
	SpaceFloor    int64
	DisableCOW    bool
	IncludedDisks map[string]*UnraidDisk
	ExcludedDisks map[string]*UnraidDisk
	CFGFile       string
}

type FSElement interface {
	GetMetadata() *Metadata
	GetSourcePath() string
	GetDestPath() string
}

type Moveable struct {
	Share      *UnraidShare
	Source     UnraidStoreable
	SourcePath string
	Dest       UnraidStoreable
	DestPath   string
	Hardlinks  []*Moveable
	Hardlink   bool
	HardlinkTo *Moveable
	Symlinks   []*Moveable
	Symlink    bool
	SymlinkTo  *Moveable
	Metadata   *Metadata
	RootDir    *RelatedDirectory
	DeepestDir *RelatedDirectory
}

func (m *Moveable) GetMetadata() *Metadata {
	return m.Metadata
}

func (m *Moveable) GetSourcePath() string {
	return m.SourcePath
}

func (m *Moveable) GetDestPath() string {
	return m.DestPath
}

type RelatedDirectory struct {
	SourcePath string
	DestPath   string
	Metadata   *Metadata
	Parent     *RelatedDirectory
	Child      *RelatedDirectory
}

func (d *RelatedDirectory) GetMetadata() *Metadata {
	return d.Metadata
}

func (d *RelatedDirectory) GetSourcePath() string {
	return d.SourcePath
}

func (d *RelatedDirectory) GetDestPath() string {
	return d.DestPath
}

type Metadata struct {
	Inode      uint64
	Perms      uint32
	UID        uint32
	GID        uint32
	AccessedAt unix.Timespec
	ModifiedAt unix.Timespec
	Size       int64
	IsDir      bool
	IsSymlink  bool
	SymlinkTo  string
}

type DiskStats struct {
	TotalSize int64
	FreeSpace int64
}

type BatchProgress struct {
	AnyProcessed       []FSElement
	DirsProcessed      []*RelatedDirectory
	MoveablesProcessed []*Moveable
	SymlinksProcessed  []*Moveable
	HardlinksProcessed []*Moveable
}
