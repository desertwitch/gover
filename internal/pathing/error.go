package pathing

import "errors"

var (
	// ErrImpossibleType is a type error ocurring when a function receives
	// a [schema.Storage] implementation that it does not support.
	ErrImpossibleType = errors.New("impossible storage type")

	ErrSourceIsRelative = errors.New("source path is relative")
	ErrNilDestination   = errors.New("destination is nil")
	ErrPathExistsOnDest = errors.New("path exists on destination")
)
