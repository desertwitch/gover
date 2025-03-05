package io

import "errors"

var (
	ErrSourceFileInUse  = errors.New("source file is currently in use")
	ErrNotEnoughSpace   = errors.New("not enough free space on destination")
	ErrHashMismatch     = errors.New("hash mismatch")
	ErrRenameExists     = errors.New("rename destination already exists")
	ErrNothingToProcess = errors.New("moveable with nothing process")
)
