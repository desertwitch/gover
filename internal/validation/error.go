package validation

import "errors"

var (
	// ErrDestMismatch occurs when the provided destination path doesn't match
	// the file system element path.
	ErrDestMismatch = errors.New("destination path mismatches destination fs element")

	// ErrDestNotConnectBase indicates that a directory's root doesn't properly
	// connect to the base of the share (destination side).
	ErrDestNotConnectBase = errors.New("related dir root does not connect to share base (dest)")

	// ErrDestPathRelative occurs when a destination path is provided as
	// relative rather than absolute.
	ErrDestPathRelative = errors.New("destination path is relative")

	// ErrHardlinkHasSublinks indicates an invalid state where a hardlink
	// contains sublinks, which shouldn't be possible.
	ErrHardlinkHasSublinks = errors.New("hardlink has sublinks")

	// ErrHardlinkSetTarget occurs when hardlink is set to false but still has a
	// target specified.
	ErrHardlinkSetTarget = errors.New("hardlink false, but has set target")

	// ErrNoDestination occurs when no destination or destination path is set.
	ErrNoDestination = errors.New("no destination or destination path")

	// ErrNoHardlinkTarget occurs when a hardlink is specified but no target
	// path is set.
	ErrNoHardlinkTarget = errors.New("no hardlink target")

	// ErrNoMetadata indicates that required metadata is missing for an
	// operation.
	ErrNoMetadata = errors.New("no metadata")

	// ErrNoRelatedDestPath occurs when a related directory is missing the
	// destination path.
	ErrNoRelatedDestPath = errors.New("no related dir destination path")

	// ErrNoRelatedMetadata indicates missing metadata for a related directory.
	ErrNoRelatedMetadata = errors.New("no related dir metadata")

	// ErrNoRelatedSourcePath occurs when a related directory is missing the
	// source path.
	ErrNoRelatedSourcePath = errors.New("no related dir source path")

	// ErrNoShareInfo occurs when share information is not set.
	ErrNoShareInfo = errors.New("no share information")

	// ErrNoSource indicates that no source or source path was set.
	ErrNoSource = errors.New("no source or source path")

	// ErrNoSymlinkTarget occurs when a symlink is specified but no target is
	// set.
	ErrNoSymlinkTarget = errors.New("no symlink target")

	// ErrRelatedDestRelative occurs when a related directory destination path
	// is relative instead of absolute.
	ErrRelatedDestRelative = errors.New("related dir destination path is relative")

	// ErrRelatedDirNotDir indicates that a path specified as a related
	// directory is not actually a directory.
	ErrRelatedDirNotDir = errors.New("related dir is not a dir")

	// ErrRelatedDirSymlink occurs when a related directory is actually a
	// symlink, which may cause unexpected behavior.
	ErrRelatedDirSymlink = errors.New("related dir is a symlink")

	// ErrRelatedSourceRelative occurs when a related directory source path is
	// relative instead of absolute.
	ErrRelatedSourceRelative = errors.New("related dir source path is relative")

	// ErrSourceMismatch occurs when the provided source path doesn't match the
	// file system element path.
	ErrSourceMismatch = errors.New("source path mismatches source fs element")

	// ErrSourceNotConnectBase indicates that a directory's root doesn't
	// properly connect to the base of the share (source side).
	ErrSourceNotConnectBase = errors.New("related dir root does not connect to share base (source)")

	// ErrSourcePathRelative occurs when a source path is provided as relative
	// rather than absolute.
	ErrSourcePathRelative = errors.New("source path is relative")

	// ErrSymlinkHasSublinks indicates an invalid state where a symlink contains
	// sublinks, which shouldn't be possible.
	ErrSymlinkHasSublinks = errors.New("symlink has sublinks")

	// ErrSymlinkSetTarget occurs when symlink is set to false but still has a
	// target specified.
	ErrSymlinkSetTarget = errors.New("symlink false, but has set target")
)
