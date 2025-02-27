package syscalls

import (
	"os"

	"golang.org/x/sys/unix"
)

type RealOS struct{}

func (RealOS) Remove(name string) error {
	return os.Remove(name)
}

func (RealOS) Readlink(name string) (string, error) {
	return os.Readlink(name)
}

func (RealOS) ReadDir(name string) ([]os.DirEntry, error) {
	return os.ReadDir(name)
}

func (RealOS) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (RealOS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (RealOS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (RealOS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

type RealUnix struct{}

func (RealUnix) Statfs(path string, buf *unix.Statfs_t) error {
	return unix.Statfs(path, buf)
}

func (RealUnix) Lstat(path string, stat *unix.Stat_t) error {
	return unix.Lstat(path, stat)
}

func (RealUnix) Link(oldpath, newpath string) error {
	return unix.Link(oldpath, newpath)
}

func (RealUnix) Symlink(oldpath, newpath string) error {
	return unix.Symlink(oldpath, newpath)
}

func (RealUnix) Mkdir(path string, mode uint32) error {
	return unix.Mkdir(path, mode)
}

func (RealUnix) Chown(path string, uid, gid int) error {
	return unix.Chown(path, uid, gid)
}

func (RealUnix) Chmod(path string, mode uint32) error {
	return unix.Chmod(path, mode)
}

func (RealUnix) Lchown(path string, uid, gid int) error {
	return unix.Lchown(path, uid, gid)
}

func (RealUnix) UtimesNano(path string, times []unix.Timespec) error {
	return unix.UtimesNano(path, times)
}
