package filesystem

import "errors"

var (
	ErrSourceIsRelative = errors.New("source path is relative")
	ErrNilDestination   = errors.New("destination is nil")
	ErrImpossibleType   = errors.New("impossible storeable type")
	ErrInvalidFileSize  = errors.New("invalid file size < 0")
	ErrInvalidStats     = errors.New("invalid stats")
	ErrPathExistsOnDest = errors.New("path exists on destination")
)
