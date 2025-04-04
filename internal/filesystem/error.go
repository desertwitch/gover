package filesystem

import "errors"

var (
	ErrImpossibleType  = errors.New("impossible storeable type")
	ErrNilDestination  = errors.New("destination is nil")
	ErrInvalidFileSize = errors.New("invalid file size < 0")
	ErrInvalidStats    = errors.New("invalid stats")
)
