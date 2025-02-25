package main

import (
	"errors"
	"fmt"
)

var (
	ErrNotAllocatable        = errors.New("no suitable destination for allocation method")
	ErrSplitDoesNotExceedLvl = errors.New("split level does not exceed max split level")

	ErrExistenceNonFatal        = errors.New("not processing as existing")
	ErrFileExistsNonFatal       = fmt.Errorf("file exists: %w", ErrExistenceNonFatal)
	ErrDirExistsNonFatal        = fmt.Errorf("directory exists: %w", ErrExistenceNonFatal)
	ErrHardlinkExistsNonFatal   = fmt.Errorf("hardlink exists: %w", ErrExistenceNonFatal)
	ErrSymlinkExistsNonFatal    = fmt.Errorf("symlink exists: %w", ErrExistenceNonFatal)
	ErrExtSymlinkExistsNonFatal = fmt.Errorf("symlink-e exists: %w", ErrExistenceNonFatal)
)
