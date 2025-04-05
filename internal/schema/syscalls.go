package schema

import (
	"os"

	"golang.org/x/sys/unix"
)

// OS is an implementation wrapping operating system functions.
type OS struct{}

// Remove wraps around [os.Remove].
func (*OS) Remove(name string) error {
	return os.Remove(name)
}

// Readlink wraps around [os.Readlink].
func (*OS) Readlink(name string) (string, error) {
	return os.Readlink(name)
}

// ReadDir wraps around [os.ReadDir].
func (*OS) ReadDir(name string) ([]os.DirEntry, error) {
	return os.ReadDir(name)
}

// Open wraps around [os.Open].
func (*OS) Open(name string) (*os.File, error) {
	return os.Open(name)
}

// OpenFile wraps around [os.OpenFile].
func (*OS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

// Stat wraps around [os.Stat].
func (*OS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Rename wraps around [os.Rename].
func (*OS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Unix is an implementation wrapping Unix operating system functions.
type Unix struct{}

// Statfs wraps around [unix.Statfs].
func (*Unix) Statfs(path string, buf *unix.Statfs_t) error {
	return unix.Statfs(path, buf)
}

// Lstat wraps around [unix.Lstat].
func (*Unix) Lstat(path string, stat *unix.Stat_t) error {
	return unix.Lstat(path, stat)
}

// Link wraps around [unix.Link].
func (*Unix) Link(oldpath, newpath string) error {
	return unix.Link(oldpath, newpath)
}

// Symlink wraps around [unix.Symlink].
func (*Unix) Symlink(oldpath, newpath string) error {
	return unix.Symlink(oldpath, newpath)
}

// Mkdir wraps around [unix.Mkdir].
func (*Unix) Mkdir(path string, mode uint32) error {
	return unix.Mkdir(path, mode)
}

// Chown wraps around [unix.Chown].
func (*Unix) Chown(path string, uid, gid int) error {
	return unix.Chown(path, uid, gid)
}

// Chmod wraps around [unix.Chmod].
func (*Unix) Chmod(path string, mode uint32) error {
	return unix.Chmod(path, mode)
}

// Lchown wraps around [unix.Lchown].
func (*Unix) Lchown(path string, uid, gid int) error {
	return unix.Lchown(path, uid, gid)
}

// UtimesNano wraps around [unix.UtimesNano].
func (*Unix) UtimesNano(path string, times []unix.Timespec) error {
	return unix.UtimesNano(path, times)
}
