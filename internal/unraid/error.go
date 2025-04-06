package unraid

import "errors"

var (
	// ErrConfPoolNotFound is an error that occurs when a configuration for a
	// [Pool] exists, but no mountpoint for the [Pool] can be found.
	ErrConfPoolNotFound = errors.New("configured pool does not exist")

	// ErrConfDiskNotFound is an error that occurs when a configuration for a
	// [Disk] exists, but no mountpoint for the [Disk] can be found.
	ErrConfDiskNotFound = errors.New("configured disk does not exist")
)
