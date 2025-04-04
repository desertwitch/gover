package pathing

import "errors"

var (
	ErrSourceIsRelative = errors.New("source path is relative")
	ErrNilDestination   = errors.New("destination is nil")
	ErrPathExistsOnDest = errors.New("path exists on destination")
)
