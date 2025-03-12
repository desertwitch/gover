package filesystem

import (
	"os"

	"golang.org/x/sys/unix"
)

type OS struct{}

func (*OS) Remove(name string) error {
	return os.Remove(name)
}

func (*OS) Readlink(name string) (string, error) {
	return os.Readlink(name)
}

func (*OS) ReadDir(name string) ([]os.DirEntry, error) {
	return os.ReadDir(name)
}

func (*OS) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (*OS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (*OS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (*OS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

type Unix struct{}

func (*Unix) Statfs(path string, buf *unix.Statfs_t) error {
	return unix.Statfs(path, buf)
}

func (*Unix) Lstat(path string, stat *unix.Stat_t) error {
	return unix.Lstat(path, stat)
}

func (*Unix) Link(oldpath, newpath string) error {
	return unix.Link(oldpath, newpath)
}

func (*Unix) Symlink(oldpath, newpath string) error {
	return unix.Symlink(oldpath, newpath)
}

func (*Unix) Mkdir(path string, mode uint32) error {
	return unix.Mkdir(path, mode)
}

func (*Unix) Chown(path string, uid, gid int) error {
	return unix.Chown(path, uid, gid)
}

func (*Unix) Chmod(path string, mode uint32) error {
	return unix.Chmod(path, mode)
}

func (*Unix) Lchown(path string, uid, gid int) error {
	return unix.Lchown(path, uid, gid)
}

func (*Unix) UtimesNano(path string, times []unix.Timespec) error {
	return unix.UtimesNano(path, times)
}
