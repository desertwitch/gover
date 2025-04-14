package io

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/desertwitch/gover/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureDirectoryStructure_Success_Nil(t *testing.T) {
	t.Parallel()

	handler := NewHandler(
		newMockFsProvider(t),
		newMockOsProvider(t),
		newMockUnixProvider(t),
	)

	m := &schema.Moveable{
		RootDir: nil,
	}

	err := handler.ensureDirectoryStructure(m, &ioReport{})

	require.NoError(t, err, "no error should occur on nil")
}

func TestEnsureDirectoryStructure_Success_NoExist(t *testing.T) {
	t.Parallel()

	fsProv := newMockFsProvider(t)
	osProv := newMockOsProvider(t)
	unixProv := newMockUnixProvider(t)
	handler := NewHandler(fsProv, osProv, unixProv)

	dir1 := &schema.Directory{
		DestPath: "/mnt/cache/dir1",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
	}

	dir2 := &schema.Directory{
		DestPath: "/mnt/cache/dir1/dir2",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
		Parent: dir1,
	}

	dir3 := &schema.Directory{
		DestPath: "/mnt/cache/dir1/dir2/dir3",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
		Parent: dir2,
	}

	dir1.Child = dir2
	dir2.Child = dir3

	osProv.EXPECT().Stat(dir1.DestPath).Return(nil, fs.ErrNotExist).Once()
	unixProv.EXPECT().Mkdir(dir1.DestPath, dir1.Metadata.Perms).Return(nil).Once()
	unixProv.EXPECT().Chown(dir1.DestPath, int(dir1.Metadata.UID), int(dir1.Metadata.GID)).Return(nil).Once()
	unixProv.EXPECT().Chmod(dir1.DestPath, dir1.Metadata.Perms).Return(nil).Once()

	osProv.EXPECT().Stat(dir2.DestPath).Return(nil, fs.ErrNotExist).Once()
	unixProv.EXPECT().Mkdir(dir2.DestPath, dir2.Metadata.Perms).Return(nil).Once()
	unixProv.EXPECT().Chown(dir2.DestPath, int(dir2.Metadata.UID), int(dir2.Metadata.GID)).Return(nil).Once()
	unixProv.EXPECT().Chmod(dir2.DestPath, dir2.Metadata.Perms).Return(nil).Once()

	osProv.EXPECT().Stat(dir3.DestPath).Return(nil, fs.ErrNotExist).Once()
	unixProv.EXPECT().Mkdir(dir3.DestPath, dir3.Metadata.Perms).Return(nil).Once()
	unixProv.EXPECT().Chown(dir3.DestPath, int(dir3.Metadata.UID), int(dir3.Metadata.GID)).Return(nil).Once()
	unixProv.EXPECT().Chmod(dir3.DestPath, dir3.Metadata.Perms).Return(nil).Once()

	ioReport := &ioReport{}
	m := &schema.Moveable{
		RootDir: dir1,
	}

	err := handler.ensureDirectoryStructure(m, ioReport)
	require.NoError(t, err, "no error should occur")

	require.Len(t, ioReport.AnyCreated, 3)
	assert.Equal(t, []fsElement{dir1, dir2, dir3}, ioReport.AnyCreated)

	require.Len(t, ioReport.DirsCreated, 3)
	assert.Equal(t, []*schema.Directory{dir1, dir2, dir3}, ioReport.DirsCreated)

	require.Len(t, ioReport.DirsWalked, 3)
	assert.Equal(t, []*schema.Directory{dir1, dir2, dir3}, ioReport.DirsWalked)

	fsProv.AssertExpectations(t)
	osProv.AssertExpectations(t)
	unixProv.AssertExpectations(t)
}

func TestEnsureDirectoryStructure_Success_AllExist(t *testing.T) {
	t.Parallel()

	fsProv := newMockFsProvider(t)
	osProv := newMockOsProvider(t)
	unixProv := newMockUnixProvider(t)
	handler := NewHandler(fsProv, osProv, unixProv)

	dir1 := &schema.Directory{
		DestPath: "/mnt/cache/dir1",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
	}

	dir2 := &schema.Directory{
		DestPath: "/mnt/cache/dir1/dir2",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
		Parent: dir1,
	}

	dir3 := &schema.Directory{
		DestPath: "/mnt/cache/dir1/dir2/dir3",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
		Parent: dir2,
	}

	dir1.Child = dir2
	dir2.Child = dir3

	osProv.EXPECT().Stat(dir1.DestPath).Return(nil, nil).Once()
	osProv.EXPECT().Stat(dir2.DestPath).Return(nil, nil).Once()
	osProv.EXPECT().Stat(dir3.DestPath).Return(nil, nil).Once()

	ioReport := &ioReport{}
	m := &schema.Moveable{
		RootDir: dir1,
	}

	err := handler.ensureDirectoryStructure(m, ioReport)
	require.NoError(t, err, "no error should occur")

	require.Empty(t, ioReport.AnyCreated)
	require.Empty(t, ioReport.DirsCreated)

	require.Len(t, ioReport.DirsWalked, 3)
	assert.Equal(t, []*schema.Directory{dir1, dir2, dir3}, ioReport.DirsWalked)

	fsProv.AssertExpectations(t)
	osProv.AssertExpectations(t)
	unixProv.AssertExpectations(t)
}

func TestEnsureDirectoryStructure_Success_Mixed(t *testing.T) {
	t.Parallel()

	fsProv := newMockFsProvider(t)
	osProv := newMockOsProvider(t)
	unixProv := newMockUnixProvider(t)
	handler := NewHandler(fsProv, osProv, unixProv)

	dir1 := &schema.Directory{
		DestPath: "/mnt/cache/dir1",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
	}

	dir2 := &schema.Directory{
		DestPath: "/mnt/cache/dir1/dir2",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
		Parent: dir1,
	}

	dir3 := &schema.Directory{
		DestPath: "/mnt/cache/dir1/dir2/dir3",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
		Parent: dir2,
	}

	dir1.Child = dir2
	dir2.Child = dir3

	osProv.EXPECT().Stat(dir1.DestPath).Return(nil, fs.ErrNotExist).Once()
	unixProv.EXPECT().Mkdir(dir1.DestPath, dir1.Metadata.Perms).Return(nil).Once()
	unixProv.EXPECT().Chown(dir1.DestPath, int(dir1.Metadata.UID), int(dir1.Metadata.GID)).Return(nil).Once()
	unixProv.EXPECT().Chmod(dir1.DestPath, dir1.Metadata.Perms).Return(nil).Once()

	osProv.EXPECT().Stat(dir2.DestPath).Return(nil, nil).Once()

	osProv.EXPECT().Stat(dir3.DestPath).Return(nil, fs.ErrNotExist).Once()
	unixProv.EXPECT().Mkdir(dir3.DestPath, dir3.Metadata.Perms).Return(nil).Once()
	unixProv.EXPECT().Chown(dir3.DestPath, int(dir3.Metadata.UID), int(dir3.Metadata.GID)).Return(nil).Once()
	unixProv.EXPECT().Chmod(dir3.DestPath, dir3.Metadata.Perms).Return(nil).Once()

	ioReport := &ioReport{}
	m := &schema.Moveable{
		RootDir: dir1,
	}

	err := handler.ensureDirectoryStructure(m, ioReport)
	require.NoError(t, err, "no error should occur")

	require.Len(t, ioReport.AnyCreated, 2)
	assert.Equal(t, []fsElement{dir1, dir3}, ioReport.AnyCreated)

	require.Len(t, ioReport.DirsCreated, 2)
	assert.Equal(t, []*schema.Directory{dir1, dir3}, ioReport.DirsCreated)

	require.Len(t, ioReport.DirsWalked, 3)
	assert.Equal(t, []*schema.Directory{dir1, dir2, dir3}, ioReport.DirsWalked)

	fsProv.AssertExpectations(t)
	osProv.AssertExpectations(t)
	unixProv.AssertExpectations(t)
}

func TestEnsureDirectoryStructure_Fail_StatError(t *testing.T) {
	t.Parallel()

	fsProv := newMockFsProvider(t)
	osProv := newMockOsProvider(t)
	unixProv := newMockUnixProvider(t)
	handler := NewHandler(fsProv, osProv, unixProv)

	dir1 := &schema.Directory{
		DestPath: "/mnt/cache/dir1",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
	}

	dir2 := &schema.Directory{
		DestPath: "/mnt/cache/dir1/dir2",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
		Parent: dir1,
	}

	dir3 := &schema.Directory{
		DestPath: "/mnt/cache/dir1/dir2/dir3",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
		Parent: dir2,
	}

	dir1.Child = dir2
	dir2.Child = dir3

	osProv.EXPECT().Stat(dir1.DestPath).Return(nil, fs.ErrNotExist).Once()
	unixProv.EXPECT().Mkdir(dir1.DestPath, dir1.Metadata.Perms).Return(nil).Once()
	unixProv.EXPECT().Chown(dir1.DestPath, int(dir1.Metadata.UID), int(dir1.Metadata.GID)).Return(nil).Once()
	unixProv.EXPECT().Chmod(dir1.DestPath, dir1.Metadata.Perms).Return(nil).Once()

	osProv.EXPECT().Stat(dir2.DestPath).Return(nil, errors.New("stat error")).Once()

	ioReport := &ioReport{}
	m := &schema.Moveable{
		RootDir: dir1,
	}

	err := handler.ensureDirectoryStructure(m, ioReport)
	require.Error(t, err, "an error should occur")

	require.Len(t, ioReport.AnyCreated, 1)
	assert.Equal(t, []fsElement{dir1}, ioReport.AnyCreated)

	require.Len(t, ioReport.DirsCreated, 1)
	assert.Equal(t, []*schema.Directory{dir1}, ioReport.DirsCreated)

	require.Len(t, ioReport.DirsWalked, 1)
	assert.Equal(t, []*schema.Directory{dir1}, ioReport.DirsWalked)

	fsProv.AssertExpectations(t)
	osProv.AssertExpectations(t)
	unixProv.AssertExpectations(t)
}

func TestEnsureDirectoryStructure_Fail_MkdirError(t *testing.T) {
	t.Parallel()

	fsProv := newMockFsProvider(t)
	osProv := newMockOsProvider(t)
	unixProv := newMockUnixProvider(t)
	handler := NewHandler(fsProv, osProv, unixProv)

	dir1 := &schema.Directory{
		DestPath: "/mnt/cache/dir1",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
	}

	dir2 := &schema.Directory{
		DestPath: "/mnt/cache/dir1/dir2",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
		Parent: dir1,
	}

	dir3 := &schema.Directory{
		DestPath: "/mnt/cache/dir1/dir2/dir3",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
		Parent: dir2,
	}

	dir1.Child = dir2
	dir2.Child = dir3

	osProv.EXPECT().Stat(dir1.DestPath).Return(nil, fs.ErrNotExist).Once()
	unixProv.EXPECT().Mkdir(dir1.DestPath, dir1.Metadata.Perms).Return(nil).Once()
	unixProv.EXPECT().Chown(dir1.DestPath, int(dir1.Metadata.UID), int(dir1.Metadata.GID)).Return(nil).Once()
	unixProv.EXPECT().Chmod(dir1.DestPath, dir1.Metadata.Perms).Return(nil).Once()

	osProv.EXPECT().Stat(dir2.DestPath).Return(nil, fs.ErrNotExist).Once()
	unixProv.EXPECT().Mkdir(dir2.DestPath, dir2.Metadata.Perms).Return(errors.New("mkdir error")).Once()

	ioReport := &ioReport{}
	m := &schema.Moveable{
		RootDir: dir1,
	}

	err := handler.ensureDirectoryStructure(m, ioReport)
	require.Error(t, err, "an error should occur")

	require.Len(t, ioReport.AnyCreated, 1)
	assert.Equal(t, []fsElement{dir1}, ioReport.AnyCreated)

	require.Len(t, ioReport.DirsCreated, 1)
	assert.Equal(t, []*schema.Directory{dir1}, ioReport.DirsCreated)

	require.Len(t, ioReport.DirsWalked, 1)
	assert.Equal(t, []*schema.Directory{dir1}, ioReport.DirsWalked)

	fsProv.AssertExpectations(t)
	osProv.AssertExpectations(t)
	unixProv.AssertExpectations(t)
}

func TestEnsureDirectoryStructure_Fail_PermError(t *testing.T) {
	t.Parallel()

	fsProv := newMockFsProvider(t)
	osProv := newMockOsProvider(t)
	unixProv := newMockUnixProvider(t)
	handler := NewHandler(fsProv, osProv, unixProv)

	dir1 := &schema.Directory{
		DestPath: "/mnt/cache/dir1",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
	}

	dir2 := &schema.Directory{
		DestPath: "/mnt/cache/dir1/dir2",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
		Parent: dir1,
	}

	dir3 := &schema.Directory{
		DestPath: "/mnt/cache/dir1/dir2/dir3",
		Metadata: &schema.Metadata{
			Perms: 0o777,
			UID:   uint32(100),
			GID:   uint32(100),
		},
		Parent: dir2,
	}

	dir1.Child = dir2
	dir2.Child = dir3

	osProv.EXPECT().Stat(dir1.DestPath).Return(nil, fs.ErrNotExist).Once()
	unixProv.EXPECT().Mkdir(dir1.DestPath, dir1.Metadata.Perms).Return(nil).Once()
	unixProv.EXPECT().Chown(dir1.DestPath, int(dir1.Metadata.UID), int(dir1.Metadata.GID)).Return(nil).Once()
	unixProv.EXPECT().Chmod(dir1.DestPath, dir1.Metadata.Perms).Return(nil).Once()

	osProv.EXPECT().Stat(dir2.DestPath).Return(nil, fs.ErrNotExist).Once()
	unixProv.EXPECT().Mkdir(dir2.DestPath, dir2.Metadata.Perms).Return(nil).Once()
	unixProv.EXPECT().Chown(dir2.DestPath, int(dir2.Metadata.UID), int(dir2.Metadata.GID)).Return(errors.New("perm error")).Once()

	ioReport := &ioReport{}
	m := &schema.Moveable{
		RootDir: dir1,
	}

	err := handler.ensureDirectoryStructure(m, ioReport)
	require.Error(t, err, "an error should occur")

	require.Len(t, ioReport.AnyCreated, 2)
	assert.Equal(t, []fsElement{dir1, dir2}, ioReport.AnyCreated)

	require.Len(t, ioReport.DirsCreated, 2)
	assert.Equal(t, []*schema.Directory{dir1, dir2}, ioReport.DirsCreated)

	require.Len(t, ioReport.DirsWalked, 2)
	assert.Equal(t, []*schema.Directory{dir1, dir2}, ioReport.DirsWalked)

	fsProv.AssertExpectations(t)
	osProv.AssertExpectations(t)
	unixProv.AssertExpectations(t)
}
