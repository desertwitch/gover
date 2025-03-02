package unraid

import (
	"fmt"
	"path/filepath"
	"regexp"
)

type UnraidArray struct {
	Disks         map[string]*UnraidDisk
	Status        string
	TurboSetting  string
	ParityRunning bool
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

// establishArray returns a pointer to an established Unraid array.
func (u *UnraidHandler) EstablishArray(disks map[string]*UnraidDisk) (*UnraidArray, error) {
	stateFile := ArrayStateFile

	configMap, err := u.ConfigOps.ReadGeneric(stateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load disk state file %s: %w", stateFile, err)
	}

	array := &UnraidArray{
		Disks:         disks,
		Status:        u.ConfigOps.MapKeyToString(configMap, StateArrayStatus),
		TurboSetting:  u.ConfigOps.MapKeyToString(configMap, StateTurboSetting),
		ParityRunning: u.ConfigOps.MapKeyToInt(configMap, StateParityPosition) > 0,
	}

	return array, nil
}

// establishDisks returns a map of pointers to established Unraid disks.
func (u *UnraidHandler) EstablishDisks() (map[string]*UnraidDisk, error) {
	basePath := BasePathMounts
	diskPattern := regexp.MustCompile(PatternDisks)

	disks := make(map[string]*UnraidDisk)

	entries, err := u.FSOps.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mounts at %s: %w", basePath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() && diskPattern.MatchString(entry.Name()) {
			disk := &UnraidDisk{
				Name:           entry.Name(),
				FSPath:         filepath.Join(basePath, entry.Name()),
				ActiveTransfer: false,
			}
			disks[disk.Name] = disk
		}
	}

	return disks, nil
}
