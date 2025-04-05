package pathing

import "errors"

var (
	// ErrImpossibleType is a type error ocurring when pathing receives a
	// [schema.Storage] implementation that it does not know to handle.
	ErrImpossibleType = errors.New("impossible storage type")

	// ErrSourceIsRelative is an error that occurs when the source path of a
	// given [schema.Moveable] is relative and not absolute.
	ErrSourceIsRelative = errors.New("source path is relative")

	// ErrNilDestination is an error that occurs when the destination of a given
	// [schema.Moveable] is nil.
	ErrNilDestination = errors.New("destination is nil")

	// ErrPathExistsOnDest is an error that occurs when the constructed
	// destination path already exists.
	ErrPathExistsOnDest = errors.New("path exists on destination")
)
