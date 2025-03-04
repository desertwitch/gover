package filesystem_test

import (
	"io/fs"
	"testing"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/filesystem/mocks"
	"github.com/desertwitch/gover/internal/unraid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEstablishPaths_FileExists(t *testing.T) {
	t.Parallel()
	mockOS := new(mocks.OsProvider)

	mockFS := &filesystem.Handler{
		OSOps:   mockOS,
		UnixOps: nil,
	}

	mockDisk := &unraid.Disk{
		Name:   "disk4",
		FSPath: "/mnt/disk4",
	}

	includedDisks := make(map[string]*unraid.Disk)
	includedDisks["disk4"] = mockDisk

	mockMoveable := &filesystem.Moveable{
		Source: &unraid.Pool{
			Name:   "cache",
			FSPath: "/mnt/cache",
		},
		SourcePath: "/mnt/cache/movies/file.mp4",
		Dest:       mockDisk,
		RootDir: &filesystem.RelatedDirectory{
			SourcePath: "/mnt/cache/movies",
		},
		Share: &unraid.Share{
			Name:          "movies",
			IncludedDisks: includedDisks,
		},
	}

	mockOS.On("Stat", mock.Anything).Return(nil, nil)

	filtered, err := mockFS.EstablishPaths([]*filesystem.Moveable{mockMoveable})

	require.NoError(t, err)
	assert.Empty(t, filtered)

	mockOS.AssertExpectations(t)
}

//nolint:dupl
func TestEstablishPaths_FileNotExits(t *testing.T) {
	t.Parallel()
	mockOS := new(mocks.OsProvider)

	mockFS := &filesystem.Handler{
		OSOps:   mockOS,
		UnixOps: nil,
	}

	mockDisk := &unraid.Disk{
		Name:   "disk4",
		FSPath: "/mnt/disk4",
	}

	includedDisks := make(map[string]*unraid.Disk)
	includedDisks["disk4"] = mockDisk

	mockMoveable := &filesystem.Moveable{
		Source: &unraid.Pool{
			Name:   "cache",
			FSPath: "/mnt/cache",
		},
		SourcePath: "/mnt/cache/movies/file.mp4",
		Dest:       mockDisk,
		RootDir: &filesystem.RelatedDirectory{
			SourcePath: "/mnt/cache/movies",
		},
		Share: &unraid.Share{
			Name:          "movies",
			IncludedDisks: includedDisks,
		},
	}

	mockOS.On("Stat", mock.Anything).Return(nil, fs.ErrNotExist)

	filtered, err := mockFS.EstablishPaths([]*filesystem.Moveable{mockMoveable})

	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.EqualValues(t, mockMoveable, filtered[0])
	assert.EqualValues(t, "/mnt/disk4/movies/file.mp4", filtered[0].DestPath)
	assert.EqualValues(t, "/mnt/disk4/movies", filtered[0].RootDir.DestPath)

	mockOS.AssertExpectations(t)
}

//nolint:dupl
func TestEstablishPaths_TrailingSlashFile(t *testing.T) {
	t.Parallel()
	mockOS := new(mocks.OsProvider)

	mockFS := &filesystem.Handler{
		OSOps:   mockOS,
		UnixOps: nil,
	}

	mockDisk := &unraid.Disk{
		Name:   "disk4",
		FSPath: "/mnt/disk4",
	}

	includedDisks := make(map[string]*unraid.Disk)
	includedDisks["disk4"] = mockDisk

	mockMoveable := &filesystem.Moveable{
		Source: &unraid.Pool{
			Name:   "cache",
			FSPath: "/mnt/cache/",
		},
		SourcePath: "/mnt/cache/movies//file.mp4",
		Dest:       mockDisk,
		RootDir: &filesystem.RelatedDirectory{
			SourcePath: "/mnt/cache/movies/",
		},
		Share: &unraid.Share{
			Name:          "movies",
			IncludedDisks: includedDisks,
		},
	}

	mockOS.On("Stat", mock.Anything).Return(nil, fs.ErrNotExist)

	filtered, err := mockFS.EstablishPaths([]*filesystem.Moveable{mockMoveable})

	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.EqualValues(t, mockMoveable, filtered[0])
	assert.EqualValues(t, "/mnt/disk4/movies/file.mp4", filtered[0].DestPath)
	assert.EqualValues(t, "/mnt/disk4/movies", filtered[0].RootDir.DestPath)

	mockOS.AssertExpectations(t)
}

//nolint:dupl
func TestEstablishPaths_TrailingSlashDir(t *testing.T) {
	t.Parallel()
	mockOS := new(mocks.OsProvider)

	mockFS := &filesystem.Handler{
		OSOps:   mockOS,
		UnixOps: nil,
	}

	mockDisk := &unraid.Disk{
		Name:   "disk4",
		FSPath: "/mnt/disk4",
	}

	includedDisks := make(map[string]*unraid.Disk)
	includedDisks["disk4"] = mockDisk

	mockMoveable := &filesystem.Moveable{
		Source: &unraid.Pool{
			Name:   "cache",
			FSPath: "/mnt/cache",
		},
		SourcePath: "/mnt/cache//movies/dir/",
		Dest:       mockDisk,
		RootDir: &filesystem.RelatedDirectory{
			SourcePath: "/mnt/cache/movies/",
		},
		Share: &unraid.Share{
			Name:          "movies",
			IncludedDisks: includedDisks,
		},
	}

	mockOS.On("Stat", mock.Anything).Return(nil, fs.ErrNotExist)

	filtered, err := mockFS.EstablishPaths([]*filesystem.Moveable{mockMoveable})

	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.EqualValues(t, mockMoveable, filtered[0])
	assert.EqualValues(t, "/mnt/disk4/movies/dir", filtered[0].DestPath)
	assert.EqualValues(t, "/mnt/disk4/movies", filtered[0].RootDir.DestPath)

	mockOS.AssertExpectations(t)
}

//nolint:dupl
func TestEstablishPaths_Unicode(t *testing.T) {
	t.Parallel()
	mockOS := new(mocks.OsProvider)

	mockFS := &filesystem.Handler{
		OSOps:   mockOS,
		UnixOps: nil,
	}

	mockDisk := &unraid.Disk{
		Name:   "disk4",
		FSPath: "/mnt/disk4",
	}

	includedDisks := make(map[string]*unraid.Disk)
	includedDisks["disk4"] = mockDisk

	mockMoveable := &filesystem.Moveable{
		Source: &unraid.Pool{
			Name:   "cache",
			FSPath: "/mnt/cache",
		},
		SourcePath: "/mnt/cache/movies/日本国",
		Dest:       mockDisk,
		RootDir: &filesystem.RelatedDirectory{
			SourcePath: "/mnt/cache/movies/",
		},
		Share: &unraid.Share{
			Name:          "movies",
			IncludedDisks: includedDisks,
		},
	}

	mockOS.On("Stat", mock.Anything).Return(nil, fs.ErrNotExist)

	filtered, err := mockFS.EstablishPaths([]*filesystem.Moveable{mockMoveable})

	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.EqualValues(t, mockMoveable, filtered[0])
	assert.EqualValues(t, "/mnt/disk4/movies/日本国", filtered[0].DestPath)
	assert.EqualValues(t, "/mnt/disk4/movies", filtered[0].RootDir.DestPath)

	mockOS.AssertExpectations(t)
}

//nolint:dupl
func TestEstablishPaths_Spaces(t *testing.T) {
	t.Parallel()
	mockOS := new(mocks.OsProvider)

	mockFS := &filesystem.Handler{
		OSOps:   mockOS,
		UnixOps: nil,
	}

	mockDisk := &unraid.Disk{
		Name:   "disk4",
		FSPath: "/mnt/disk4",
	}

	includedDisks := make(map[string]*unraid.Disk)
	includedDisks["disk4"] = mockDisk

	mockMoveable := &filesystem.Moveable{
		Source: &unraid.Pool{
			Name:   "cache",
			FSPath: "/mnt/cache",
		},
		SourcePath: "/mnt/cache/movi es/日本国",
		Dest:       mockDisk,
		RootDir: &filesystem.RelatedDirectory{
			SourcePath: "/mnt/cache/movi es/",
		},
		Share: &unraid.Share{
			Name:          "movi es",
			IncludedDisks: includedDisks,
		},
	}

	mockOS.On("Stat", mock.Anything).Return(nil, fs.ErrNotExist)

	filtered, err := mockFS.EstablishPaths([]*filesystem.Moveable{mockMoveable})

	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.EqualValues(t, mockMoveable, filtered[0])
	assert.EqualValues(t, "/mnt/disk4/movi es/日本国", filtered[0].DestPath)
	assert.EqualValues(t, "/mnt/disk4/movi es", filtered[0].RootDir.DestPath)

	mockOS.AssertExpectations(t)
}
