package unraid

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFindPool tests the helper function [findPool].
func TestFindPool(t *testing.T) {
	t.Parallel()

	knownPools := map[string]*Pool{
		"pool1": {Name: "pool1", FSPath: "/mnt/pool1"},
	}

	t.Run("Success_EmptyPoolName", func(t *testing.T) {
		pool, err := findPool("", knownPools)
		require.NoError(t, err, "expected no error when poolName is empty")
		assert.Nil(t, pool, "expected nil pool when poolName is empty")
	})

	t.Run("Success_PoolFound", func(t *testing.T) {
		pool, err := findPool("pool1", knownPools)
		require.NoError(t, err, "expected no error when pool is found")
		require.NotNil(t, pool, "expected a non-nil pool when found")
		assert.Equal(t, "pool1", pool.Name, "expected pool name to match")
	})

	t.Run("Fail_PoolNotFound", func(t *testing.T) {
		pool, err := findPool("pool2", knownPools)
		require.Error(t, err, "expected error when pool is not found")
		assert.Nil(t, pool, "expected nil pool when pool is not found")
		require.ErrorIs(t, err, ErrConfPoolNotFound, "error should be or wrap ErrConfPoolNotFound")
		assert.Contains(t, err.Error(), "pool2", "error message should contain pool name")
	})
}

// TestFindDisks tests the helper function [findDisks].
func TestFindDisks(t *testing.T) {
	t.Parallel()

	knownDisks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
		"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
	}

	t.Run("Success_EmptyDiskName", func(t *testing.T) {
		disks, err := findDisks("", knownDisks)
		require.NoError(t, err, "expected no error when diskNames is empty")
		assert.Nil(t, disks, "expected nil map when diskNames is empty")
	})

	t.Run("Success_AllDisksFound", func(t *testing.T) {
		disks, err := findDisks("disk1,disk2", knownDisks)
		require.NoError(t, err, "expected no error when all disks are found")
		require.NotNil(t, disks, "expected non-nil map when disks are found")
		assert.Len(t, disks, 2, "expected two disks in result")
		assert.Contains(t, disks, "disk1", "expected disk1 to be found")
		assert.Contains(t, disks, "disk2", "expected disk2 to be found")
	})

	t.Run("Fail_DiskNotFound", func(t *testing.T) {
		disks, err := findDisks("disk1,disk3", knownDisks)
		require.Error(t, err, "expected error when a disk is not found")
		assert.Nil(t, disks, "expected nil map when a disk is not found")
		require.ErrorIs(t, err, ErrConfDiskNotFound, "error should be or wrap ErrConfDiskNotFound")
		assert.Contains(t, err.Error(), "disk3", "error message should contain missing disk name")
	})
}
