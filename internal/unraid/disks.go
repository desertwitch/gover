package unraid

import (
	"fmt"
	"path/filepath"
	"regexp"
)

type Disk struct {
	Name   string
	FSPath string
}

func (d *Disk) IsDisk() bool {
	return true
}

func (d *Disk) GetName() string {
	return d.Name
}

func (d *Disk) GetFSPath() string {
	return d.FSPath
}

// establishDisks returns a map of pointers to established Unraid disks.
func (u *Handler) establishDisks() (map[string]*Disk, error) {
	basePath := BasePathMounts
	diskPattern := regexp.MustCompile(PatternDisks)

	disks := make(map[string]*Disk)

	entries, err := u.FSOps.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("(unraid-array) failed to readdir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && diskPattern.MatchString(entry.Name()) {
			disk := &Disk{
				Name:   entry.Name(),
				FSPath: filepath.Join(basePath, entry.Name()),
			}
			disks[disk.Name] = disk
		}
	}

	return disks, nil
}
