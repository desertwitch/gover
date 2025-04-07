package unraid

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/desertwitch/gover/internal/unraid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeDirEntry implements os.DirEntry for testing.
type fakeDirEntry struct {
	name  string
	isDir bool
}

func (f fakeDirEntry) Name() string { return f.name }
func (f fakeDirEntry) IsDir() bool  { return f.isDir }
func (f fakeDirEntry) Type() os.FileMode {
	if f.isDir {
		return os.ModeDir
	}

	return 0
}
func (f fakeDirEntry) Info() (os.FileInfo, error) { return nil, nil } //nolint: nilnil

// TestEstablishDisks_Success simulates success in establishing the disks.
func TestEstablishDisks_Success(t *testing.T) {
	t.Parallel()

	osMock := mocks.NewOsProvider(t)

	handler := &Handler{
		osHandler: osMock,
	}

	fakeEntries := []os.DirEntry{
		fakeDirEntry{name: "disk1", isDir: true},   // Should match.
		fakeDirEntry{name: "disk2", isDir: true},   // Should match.
		fakeDirEntry{name: "floppy1", isDir: true}, // Does not match the regex.
		fakeDirEntry{name: "disk3", isDir: false},  // Not a directory.
	}

	osMock.On("ReadDir", BasePathMounts).Return(fakeEntries, nil)

	disks, err := handler.establishDisks()
	require.NoError(t, err, "unexpected error from establishDisks")

	expectedDiskNames := []string{"disk1", "disk2"}
	assert.Len(t, disks, len(expectedDiskNames), "unexpected number of disks found")

	for _, name := range expectedDiskNames {
		disk, ok := disks[name]
		assert.True(t, ok, "expected disk %s not found", name)
		expectedFSPath := filepath.Join(BasePathMounts, name)
		assert.Equal(t, expectedFSPath, disk.FSPath, "disk %s FSPath mismatch", name)
	}

	osMock.AssertExpectations(t)
}

// TestEstablishDisks_ReadDirError simulates an error reading the mount base.
func TestEstablishDisks_Fail_ReadDirError(t *testing.T) {
	t.Parallel()

	osMock := mocks.NewOsProvider(t)
	handler := &Handler{
		osHandler: osMock,
	}

	readErr := errors.New("read dir error")
	osMock.On("ReadDir", BasePathMounts).Return(nil, readErr)

	disks, err := handler.establishDisks()
	require.Error(t, err, "expected an error from establishDisks")
	assert.Nil(t, disks, "expected no disks when ReadDir errors")
	assert.Contains(t, err.Error(), "failed to readdir", "error message should indicate readdir failure")

	osMock.AssertExpectations(t)
}
