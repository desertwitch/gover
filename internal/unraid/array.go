package unraid

import (
	"fmt"
	"path/filepath"
	"regexp"
)

type Array struct {
	Disks         map[string]*Disk
	Status        string
	TurboSetting  string
	ParityRunning bool
}

type Disk struct {
	Name           string
	FSPath         string
	ActiveTransfer bool
}

func (d *Disk) GetName() string {
	return d.Name
}

func (d *Disk) GetFSPath() string {
	return d.FSPath
}

func (d *Disk) IsActiveTransfer() bool {
	return d.ActiveTransfer
}

func (d *Disk) SetActiveTransfer(active bool) {
	d.ActiveTransfer = active
}

// establishArray returns a pointer to an established Unraid array.
func (u *Handler) EstablishArray(disks map[string]*Disk) (*Array, error) {
	stateFile := ArrayStateFile

	configMap, err := u.ConfigOps.ReadGeneric(stateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load disk state file %s: %w", stateFile, err)
	}

	array := &Array{
		Disks:         disks,
		Status:        u.ConfigOps.MapKeyToString(configMap, StateArrayStatus),
		TurboSetting:  u.ConfigOps.MapKeyToString(configMap, StateTurboSetting),
		ParityRunning: u.ConfigOps.MapKeyToInt(configMap, StateParityPosition) > 0,
	}

	return array, nil
}

// establishDisks returns a map of pointers to established Unraid disks.
func (u *Handler) EstablishDisks() (map[string]*Disk, error) {
	basePath := BasePathMounts
	diskPattern := regexp.MustCompile(PatternDisks)

	disks := make(map[string]*Disk)

	entries, err := u.FSOps.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mounts at %s: %w", basePath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() && diskPattern.MatchString(entry.Name()) {
			disk := &Disk{
				Name:           entry.Name(),
				FSPath:         filepath.Join(basePath, entry.Name()),
				ActiveTransfer: false,
			}
			disks[disk.Name] = disk
		}
	}

	return disks, nil
}
