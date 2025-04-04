package filesystem

import "errors"

var (
	// ErrImpossibleType is a type error ocurring when a function receives
	// a [schema.Storage] implementation that it does not support.
	ErrImpossibleType = errors.New("impossible storage type")

	// ErrNilDestination is an error that occurs when a [schema.Moveable]
	// destination is attempted to be accessed but is in fact nil.
	ErrNilDestination = errors.New("destination is nil")

	// ErrInvalidFileSize is an error that occurs when a given filesize is
	// smaller than 0 and impossible to handle in the respective function.
	ErrInvalidFileSize = errors.New("invalid file size < 0")
)
