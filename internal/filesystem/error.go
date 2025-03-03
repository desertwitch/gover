package filesystem

import "errors"

var (
	ErrSourceIsRelative = errors.New("source path is relative")
	ErrNilDestination   = errors.New("destination is nil")
	ErrNilDirRoot       = errors.New("dir root is nil")
	ErrImpossibleType   = errors.New("impossible storeable type")
	ErrInvalidFileSize  = errors.New("invalid file size < 0")
	ErrInvalidStats     = errors.New("invalid stats")
)
