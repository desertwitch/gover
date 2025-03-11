package unraid

import (
	"fmt"

	"github.com/desertwitch/gover/internal/generic/storage"
)

type Array struct {
	Disks         map[string]*Disk
	Status        string
	TurboSetting  string
	ParityRunning bool
}

func (a *Array) GetDisks() map[string]storage.Disk {
	if a.Disks == nil {
		return nil
	}

	disks := make(map[string]storage.Disk)
	for k, v := range a.Disks {
		if v != nil {
			disks[k] = v
		}
	}

	return disks
}

// establishArray returns a pointer to an established Unraid array.
func (u *Handler) establishArray(disks map[string]*Disk) (*Array, error) {
	stateFile := ArrayStateFile

	configMap, err := u.ConfigHandler.ReadGeneric(stateFile)
	if err != nil {
		return nil, fmt.Errorf("(unraid-array) failed to load array state file: %w", err)
	}

	array := &Array{
		Disks:         disks,
		Status:        u.ConfigHandler.MapKeyToString(configMap, StateArrayStatus),
		TurboSetting:  u.ConfigHandler.MapKeyToString(configMap, StateTurboSetting),
		ParityRunning: u.ConfigHandler.MapKeyToInt(configMap, StateParityPosition) > 0,
	}

	return array, nil
}
