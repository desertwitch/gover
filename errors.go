package main

import "errors"

var (
	ErrSplitDoesNotExceedLvl = errors.New("split level does not exceed max split level")
)
