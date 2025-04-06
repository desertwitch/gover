package unraid

import (
	"fmt"
	"path/filepath"
	"regexp"
)

// Disk is an Unraid disk, as part of an Unraid [Array].
type Disk struct {
	Name   string
	FSPath string
}

// IsDisk is a type identifier.
func (d *Disk) IsDisk() bool {
	return true
}

// GetName returns the disk name.
func (d *Disk) GetName() string {
	return d.Name
}

// GetFSPath returns an absolute filesystem path to the disk's mountpoint.
func (d *Disk) GetFSPath() string {
	return d.FSPath
}

// establishDisks returns a map (map[diskName]*Disk) to all Unraid [Disk]. It is
// the principal method for reading all disk information from the system.
func (u *Handler) establishDisks() (map[string]*Disk, error) {
	basePath := BasePathMounts
	diskPattern := regexp.MustCompile(PatternDisks)

	disks := make(map[string]*Disk)

	entries, err := u.osHandler.ReadDir(basePath)
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
