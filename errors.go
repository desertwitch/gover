package main

import "errors"

var (
	ErrNotAllocatable        = errors.New("no suitable destination for allocation method")
	ErrSplitDoesNotExceedLvl = errors.New("split level does not exceed max split level")
)
