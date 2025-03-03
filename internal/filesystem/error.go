package filesystem

import "errors"

var (
	ErrNilDestination  = errors.New("destination is nil")
	ErrNilDirRoot      = errors.New("dir root is nil")
	ErrImpossibleType  = errors.New("impossible storeable type")
	ErrInvalidFileSize = errors.New("invalid file size < 0")
	ErrInvalidStats    = errors.New("invalid stats")
)
