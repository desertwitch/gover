package io

import "errors"

var (
	// ErrSourceFileInUse is an error that occurs when the source file is already in use.
	ErrSourceFileInUse = errors.New("source file is currently in use")

	// ErrNotEnoughSpace is an error that occurs when there is not enough free space to
	// take the to be moved file on the target disk.
	ErrNotEnoughSpace = errors.New("not enough free space on destination")

	// ErrHashMismatch is an error that occurs when there is a source/destination hash
	// mismatch, this usually means that there are underlying transfer/hardware issues.
	ErrHashMismatch = errors.New("hash mismatch")

	// ErrRenameExists is an error that occurs when the intermediate file is to be renamed
	// to its final filename, but that final filename already exists on the target disk.
	ErrRenameExists = errors.New("rename destination already exists")

	// ErrNothingToProcess is a type error that occurs when a [schema.Moveable] is not
	// of a known type and the respective IO functions do not know how to process it.
	ErrNothingToProcess = errors.New("moveable with nothing process")
)
