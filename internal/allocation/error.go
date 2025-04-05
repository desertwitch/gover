package allocation

import (
	"errors"
)

var (
	// ErrNotAllocatable is an error that is returned when no allocation is
	// possible.
	ErrNotAllocatable = errors.New("no suitable destination for allocation method")

	// ErrSplitDoesNotExceedLvl is a sentinel error that is returned for
	// situational error handling, it is to exclude [schema.Moveable] not
	// relevant to split level.
	ErrSplitDoesNotExceedLvl = errors.New("split level does not exceed max split level")

	// ErrNoAllocationMethod is an error that should not be possible under
	// normal circumstances, it is returned for misconfigurations and non
	// supported allocation methods.
	ErrNoAllocationMethod = errors.New("no allocation method given in configuration")

	// ErrNoDiskStats is returned when no disk statistics were able to be
	// obtained, it is typically only seen during severe malfunctions of the
	// operating system.
	ErrNoDiskStats = errors.New("failed getting stats for any disk")

	// ErrCalcSplitLvlZero is returned when the split level calculated on a
	// moveable is zero, which should never be possible since we are operating
	// on full paths.
	ErrCalcSplitLvlZero = errors.New("calc split level of zero")
)
