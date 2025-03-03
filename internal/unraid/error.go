package unraid

import "errors"

var (
	ErrConfPoolNotFound = errors.New("configured pool does not exist")
	ErrConfDiskNotFound = errors.New("configured disk does not exist")
)
