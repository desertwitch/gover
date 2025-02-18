package main

import (
	"golang.org/x/sys/unix"
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

type Moveable struct {
	Share      *UnraidShare
	Path       string
	Source     UnraidStoreable
	Dest       UnraidStoreable
	Hardlinks  []*Moveable
	Hardlink   bool
	HardlinkTo *Moveable
	Symlinks   []*Moveable
	Symlink    bool
	SymlinkTo  *Moveable
	Metadata   *Metadata
	Parent     *RelatedDirectory
	RootDir    *RelatedDirectory
}

type RelatedDirectory struct {
	Path         string
	RelativePath string
	Metadata     *Metadata
	Parent       *RelatedDirectory
	Child        *RelatedDirectory
}

type Metadata struct {
	Inode      uint64
	Perms      uint32
	UID        uint32
	GID        uint32
	CreatedAt  unix.Timespec
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
