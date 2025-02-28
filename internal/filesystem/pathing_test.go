package filesystem_test

import (
	"io/fs"
	"testing"

	"github.com/desertwitch/gover/internal/filesystem"
	"github.com/desertwitch/gover/internal/filesystem/mocks"
	"github.com/desertwitch/gover/internal/unraid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEstablishPaths_FileExists(t *testing.T) {
	mockOS := new(mocks.OsProvider)

	mockFS := &filesystem.FileHandler{
		OSOps:   mockOS,
		UnixOps: nil,
	}

	mockDisk := &unraid.UnraidDisk{
		Name:   "disk4",
		FSPath: "/mnt/disk4",
	}

	includedDisks := make(map[string]*unraid.UnraidDisk)
	includedDisks["disk4"] = mockDisk

	mockMoveable := &filesystem.Moveable{
		Source: &unraid.UnraidPool{
			Name:   "cache",
			FSPath: "/mnt/cache",
		},
		SourcePath: "/mnt/cache/movies/file.mp4",
		Dest:       mockDisk,
		RootDir: &filesystem.RelatedDirectory{
			SourcePath: "/mnt/cache/movies",
		},
		Share: &unraid.UnraidShare{
			Name:          "movies",
			IncludedDisks: includedDisks,
		},
	}

	mockOS.On("Stat", mock.Anything).Return(nil, nil)

	filtered, err := mockFS.EstablishPaths([]*filesystem.Moveable{mockMoveable})

	assert.NoError(t, err)
	assert.Len(t, filtered, 0)

	mockOS.AssertExpectations(t)
}

func TestEstablishPaths_FileNotExits(t *testing.T) {
	mockOS := new(mocks.OsProvider)

	mockFS := &filesystem.FileHandler{
		OSOps:   mockOS,
		UnixOps: nil,
	}

	mockDisk := &unraid.UnraidDisk{
		Name:   "disk4",
		FSPath: "/mnt/disk4",
	}

	includedDisks := make(map[string]*unraid.UnraidDisk)
	includedDisks["disk4"] = mockDisk

	mockMoveable := &filesystem.Moveable{
		Source: &unraid.UnraidPool{
			Name:   "cache",
			FSPath: "/mnt/cache",
		},
		SourcePath: "/mnt/cache/movies/file.mp4",
		Dest:       mockDisk,
		RootDir: &filesystem.RelatedDirectory{
			SourcePath: "/mnt/cache/movies",
		},
		Share: &unraid.UnraidShare{
			Name:          "movies",
			IncludedDisks: includedDisks,
		},
	}

	mockOS.On("Stat", mock.Anything).Return(nil, fs.ErrNotExist)

	filtered, err := mockFS.EstablishPaths([]*filesystem.Moveable{mockMoveable})

	assert.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.EqualValues(t, filtered[0], mockMoveable)
	assert.EqualValues(t, filtered[0].DestPath, "/mnt/disk4/movies/file.mp4")
	assert.EqualValues(t, filtered[0].RootDir.DestPath, "/mnt/disk4/movies")

	mockOS.AssertExpectations(t)
}

func TestEstablishPaths_TrailingSlashFile(t *testing.T) {
	mockOS := new(mocks.OsProvider)

	mockFS := &filesystem.FileHandler{
		OSOps:   mockOS,
		UnixOps: nil,
	}

	mockDisk := &unraid.UnraidDisk{
		Name:   "disk4",
		FSPath: "/mnt/disk4",
	}

	includedDisks := make(map[string]*unraid.UnraidDisk)
	includedDisks["disk4"] = mockDisk

	mockMoveable := &filesystem.Moveable{
		Source: &unraid.UnraidPool{
			Name:   "cache",
			FSPath: "/mnt/cache/",
		},
		SourcePath: "/mnt/cache/movies//file.mp4",
		Dest:       mockDisk,
		RootDir: &filesystem.RelatedDirectory{
			SourcePath: "/mnt/cache/movies/",
		},
		Share: &unraid.UnraidShare{
			Name:          "movies",
			IncludedDisks: includedDisks,
		},
	}

	mockOS.On("Stat", mock.Anything).Return(nil, fs.ErrNotExist)

	filtered, err := mockFS.EstablishPaths([]*filesystem.Moveable{mockMoveable})

	assert.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.EqualValues(t, filtered[0], mockMoveable)
	assert.EqualValues(t, filtered[0].DestPath, "/mnt/disk4/movies/file.mp4")
	assert.EqualValues(t, filtered[0].RootDir.DestPath, "/mnt/disk4/movies")

	mockOS.AssertExpectations(t)
}

func TestEstablishPaths_TrailingSlashDir(t *testing.T) {
	mockOS := new(mocks.OsProvider)

	mockFS := &filesystem.FileHandler{
		OSOps:   mockOS,
		UnixOps: nil,
	}

	mockDisk := &unraid.UnraidDisk{
		Name:   "disk4",
		FSPath: "/mnt/disk4",
	}

	includedDisks := make(map[string]*unraid.UnraidDisk)
	includedDisks["disk4"] = mockDisk

	mockMoveable := &filesystem.Moveable{
		Source: &unraid.UnraidPool{
			Name:   "cache",
			FSPath: "/mnt/cache",
		},
		SourcePath: "/mnt/cache//movies/dir/",
		Dest:       mockDisk,
		RootDir: &filesystem.RelatedDirectory{
			SourcePath: "/mnt/cache/movies/",
		},
		Share: &unraid.UnraidShare{
			Name:          "movies",
			IncludedDisks: includedDisks,
		},
	}

	mockOS.On("Stat", mock.Anything).Return(nil, fs.ErrNotExist)

	filtered, err := mockFS.EstablishPaths([]*filesystem.Moveable{mockMoveable})

	assert.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.EqualValues(t, filtered[0], mockMoveable)
	assert.EqualValues(t, filtered[0].DestPath, "/mnt/disk4/movies/dir")
	assert.EqualValues(t, filtered[0].RootDir.DestPath, "/mnt/disk4/movies")

	mockOS.AssertExpectations(t)
}

func TestEstablishPaths_Unicode(t *testing.T) {
	mockOS := new(mocks.OsProvider)

	mockFS := &filesystem.FileHandler{
		OSOps:   mockOS,
		UnixOps: nil,
	}

	mockDisk := &unraid.UnraidDisk{
		Name:   "disk4",
		FSPath: "/mnt/disk4",
	}

	includedDisks := make(map[string]*unraid.UnraidDisk)
	includedDisks["disk4"] = mockDisk

	mockMoveable := &filesystem.Moveable{
		Source: &unraid.UnraidPool{
			Name:   "cache",
			FSPath: "/mnt/cache",
		},
		SourcePath: "/mnt/cache/movies/日本国",
		Dest:       mockDisk,
		RootDir: &filesystem.RelatedDirectory{
			SourcePath: "/mnt/cache/movies/",
		},
		Share: &unraid.UnraidShare{
			Name:          "movies",
			IncludedDisks: includedDisks,
		},
	}

	mockOS.On("Stat", mock.Anything).Return(nil, fs.ErrNotExist)

	filtered, err := mockFS.EstablishPaths([]*filesystem.Moveable{mockMoveable})

	assert.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.EqualValues(t, filtered[0], mockMoveable)
	assert.EqualValues(t, filtered[0].DestPath, "/mnt/disk4/movies/日本国")
	assert.EqualValues(t, filtered[0].RootDir.DestPath, "/mnt/disk4/movies")

	mockOS.AssertExpectations(t)
}

func TestEstablishPaths_Spaces(t *testing.T) {
	mockOS := new(mocks.OsProvider)

	mockFS := &filesystem.FileHandler{
		OSOps:   mockOS,
		UnixOps: nil,
	}

	mockDisk := &unraid.UnraidDisk{
		Name:   "disk4",
		FSPath: "/mnt/disk4",
	}

	includedDisks := make(map[string]*unraid.UnraidDisk)
	includedDisks["disk4"] = mockDisk

	mockMoveable := &filesystem.Moveable{
		Source: &unraid.UnraidPool{
			Name:   "cache",
			FSPath: "/mnt/cache",
		},
		SourcePath: "/mnt/cache/movi es/日本国",
		Dest:       mockDisk,
		RootDir: &filesystem.RelatedDirectory{
			SourcePath: "/mnt/cache/movi es/",
		},
		Share: &unraid.UnraidShare{
			Name:          "movi es",
			IncludedDisks: includedDisks,
		},
	}

	mockOS.On("Stat", mock.Anything).Return(nil, fs.ErrNotExist)

	filtered, err := mockFS.EstablishPaths([]*filesystem.Moveable{mockMoveable})

	assert.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.EqualValues(t, filtered[0], mockMoveable)
	assert.EqualValues(t, filtered[0].DestPath, "/mnt/disk4/movi es/日本国")
	assert.EqualValues(t, filtered[0].RootDir.DestPath, "/mnt/disk4/movi es")

	mockOS.AssertExpectations(t)
}
