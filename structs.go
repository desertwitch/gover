package main

import (
	"os"
	"syscall"
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
	SpaceFloor    int
	DisableCOW    bool
	IncludedDisks map[string]*UnraidDisk
	ExcludedDisks map[string]*UnraidDisk
	CFGFile       string
}

type Moveable struct {
	Share         *UnraidShare
	Path          string
	Source        UnraidStoreable
	Dest          UnraidStoreable
	Hardlink      bool
	HardlinkTo    *Moveable
	Symlink       bool
	SymlinkTo     *Moveable
	Metadata      *Metadata
	ParentDirs    map[string]*Metadata
	InternalLinks []*Moveable
}

type Metadata struct {
	Inode       uint64
	Permissions os.FileMode
	UID         uint32
	GID         uint32
	CreatedAt   syscall.Timespec
	ModifiedAt  syscall.Timespec
	Size        int64
	IsDir       bool
	IsSymlink   bool
	SymlinkTo   string
}
