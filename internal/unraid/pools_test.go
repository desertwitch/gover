package unraid

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEstablishPools_Success verifies that establishPools returns a map of
// pools, when the configuration directory exists and valid pool config files
// are found.
func TestEstablishPools_Success(t *testing.T) {
	t.Parallel()

	fsMock := newMockFsProvider(t)
	osMock := newMockOsProvider(t)

	handler := &Handler{
		fsHandler: fsMock,
		osHandler: osMock,
	}

	fsMock.On("Exists", ConfigDirPools).Return(true, nil)

	fakeFiles := []os.DirEntry{
		fakeDirEntry{name: "pool1.cfg", isDir: false},   // valid pool file
		fakeDirEntry{name: "notpool.txt", isDir: false}, // ignored
		fakeDirEntry{name: "pool2.cfg", isDir: false},   // valid pool file
		fakeDirEntry{name: "folder", isDir: true},       // ignored
	}
	osMock.On("ReadDir", ConfigDirPools).Return(fakeFiles, nil)

	fsMock.On("Exists", filepath.Join("/mnt", "pool1")).Return(true, nil)
	fsMock.On("Exists", filepath.Join("/mnt", "pool2")).Return(true, nil)

	pools, err := handler.establishPools()
	require.NoError(t, err, "establishPools should not return an error")
	require.NotNil(t, pools, "returned pools map should not be nil")
	assert.Len(t, pools, 2, "expected 2 pools to be returned")

	pool1, ok := pools["pool1"]
	require.True(t, ok, "expected pool1 not found")
	assert.Equal(t, "pool1", pool1.Name)
	assert.Equal(t, filepath.Join("/mnt", "pool1"), pool1.FSPath)

	pool2, ok := pools["pool2"]
	require.True(t, ok, "expected pool2 not found")
	assert.Equal(t, "pool2", pool2.Name)
	assert.Equal(t, filepath.Join("/mnt", "pool2"), pool2.FSPath)

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
}

// TestEstablishPools_Fail_ConfigDirDoesNotExist simulates the scenario where
// the pools configuration directory does not exist.
func TestEstablishPools_Fail_ConfigDirDoesNotExist(t *testing.T) {
	t.Parallel()

	fsMock := newMockFsProvider(t)
	osMock := newMockOsProvider(t)
	handler := &Handler{
		fsHandler: fsMock,
		osHandler: osMock,
	}

	fsMock.On("Exists", ConfigDirPools).Return(false, nil)

	pools, err := handler.establishPools()
	require.Error(t, err, "an error was expected when the config directory does not exist")
	assert.Nil(t, pools, "expected pools to be nil when the config directory is missing")
	assert.Contains(t, err.Error(), "config dir does not exist", "error message should indicate missing config directory")

	fsMock.AssertExpectations(t)
}

// TestEstablishPools_Fail_ReadDirError simulates an error while reading the
// pools configuration directory.
func TestEstablishPools_Fail_ReadDirError(t *testing.T) {
	t.Parallel()

	fsMock := newMockFsProvider(t)
	osMock := newMockOsProvider(t)
	handler := &Handler{
		fsHandler: fsMock,
		osHandler: osMock,
	}

	fsMock.On("Exists", ConfigDirPools).Return(true, nil)

	readErr := errors.New("read dir error")
	osMock.On("ReadDir", ConfigDirPools).Return(nil, readErr)

	pools, err := handler.establishPools()
	require.Error(t, err, "an error was expected due to ReadDir failure")
	assert.Nil(t, pools, "expected pools to be nil when ReadDir errors")
	assert.Contains(t, err.Error(), "failed to readdir", "error message should mention readdir failure")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
}

// TestEstablishPools_Fail_MountpointDoesNotExist simulates a scenario where a
// pool's mountpoint is missing.
func TestEstablishPools_Fail_MountpointDoesNotExist(t *testing.T) {
	t.Parallel()

	fsMock := newMockFsProvider(t)
	osMock := newMockOsProvider(t)
	handler := &Handler{
		fsHandler: fsMock,
		osHandler: osMock,
	}

	fsMock.On("Exists", ConfigDirPools).Return(true, nil)

	fakeFiles := []os.DirEntry{
		fakeDirEntry{name: "pool1.cfg", isDir: false},
	}
	osMock.On("ReadDir", ConfigDirPools).Return(fakeFiles, nil)

	fsMock.On("Exists", filepath.Join("/mnt", "pool1")).Return(false, nil)

	pools, err := handler.establishPools()
	require.Error(t, err, "an error was expected when the mountpoint does not exist")
	assert.Nil(t, pools, "expected pools to be nil when the mountpoint is missing")
	assert.Contains(t, err.Error(), "mountpoint does not exist", "error message should indicate missing mountpoint")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
}
