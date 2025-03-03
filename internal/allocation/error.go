package allocation

import (
	"errors"
)

var (
	ErrNotAllocatable        = errors.New("no suitable destination for allocation method")
	ErrSplitDoesNotExceedLvl = errors.New("split level does not exceed max split level")
	ErrNoAllocationMethod    = errors.New("no allocation method given in configuration")
	ErrNoDiskStats           = errors.New("failed getting stats for any disk")
	ErrCalcSplitLvlZero      = errors.New("calc split level of zero")
)
