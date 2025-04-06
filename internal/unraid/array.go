package unraid

import (
	"fmt"
)

// Array is an Unraid array, containing [Disk]. It is part of an Unraid [System]
// and contains all relevant information related to the array.
type Array struct {
	Disks         map[string]*Disk
	Status        string
	TurboSetting  string
	ParityRunning bool
}

// establishArray returns a pointer to an established Unraid [Array]. It is the
// principal method to retrieve all array state information from the system.
func (u *Handler) establishArray(disks map[string]*Disk) (*Array, error) {
	stateFile := ArrayStateFile

	configMap, err := u.configHandler.ReadGeneric(stateFile)
	if err != nil {
		return nil, fmt.Errorf("(unraid-array) failed to load array state file: %w", err)
	}

	array := &Array{
		Disks:         disks,
		Status:        u.configHandler.MapKeyToString(configMap, StateArrayStatus),
		TurboSetting:  u.configHandler.MapKeyToString(configMap, StateTurboSetting),
		ParityRunning: u.configHandler.MapKeyToInt(configMap, StateParityPosition) > 0,
	}

	return array, nil
}
