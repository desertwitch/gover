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

// TestEstablishShares_Success verifies that a valid share configuration file is
// processed correctly.
func TestEstablishShares_Success(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)

	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "",
		SettingGlobalShareExcludes: "",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("")

	shareFilePath := filepath.Join(ConfigDirShares, "share1.cfg")
	shareConfigMap := map[string]string{}

	configMock.On("ReadGeneric", shareFilePath).Return(shareConfigMap, nil)
	configMock.On("MapKeyToString", shareConfigMap, SettingShareUseCache).Return("yes")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool).Return("cache")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool2).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCOW).Return("auto")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareAllocator).Return("highwater")
	configMock.On("MapKeyToInt", shareConfigMap, SettingShareSplitLevel).Return(0)
	configMock.On("MapKeyToUInt64", shareConfigMap, SettingShareFloor).Return(uint64(10000000))
	configMock.On("MapKeyToString", shareConfigMap, SettingShareIncludeDisks).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareExcludeDisks).Return("")

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
	}

	pools := map[string]*Pool{
		"cache": {Name: "cache", FSPath: "/mnt/cache"},
	}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}

	shares, err := handler.establishShares(disks, pools)
	require.NoError(t, err, "establishShares should not return an error")
	require.NotNil(t, shares, "shares map should not be nil")

	share, ok := shares["share1"]
	require.True(t, ok, "share1 should be present")
	assert.Equal(t, "share1", share.Name)
	assert.Equal(t, "yes", share.UseCache)
	require.NotNil(t, share.CachePool, "CachePool should not be nil")
	assert.Equal(t, "cache", share.CachePool.Name)
	assert.Nil(t, share.CachePool2, "CachePool2 should be nil")
	assert.Equal(t, "highwater", share.Allocator)
	assert.False(t, share.DisableCOW)
	assert.Equal(t, 0, share.SplitLevel)
	assert.Equal(t, uint64(10000000), share.SpaceFloor)
	assert.Equal(t, disks, share.GetIncludedDisks(), "included disks should match all disks")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Success_NoIncludedDisks verifies that a valid share
// configuration file is processed correctly, even when there are no included
// disks.
func TestEstablishShares_Success_NoIncludedDisks(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)

	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "",
		SettingGlobalShareExcludes: "",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("")

	shareFilePath := filepath.Join(ConfigDirShares, "share1.cfg")
	shareConfigMap := map[string]string{}

	configMock.On("ReadGeneric", shareFilePath).Return(shareConfigMap, nil)
	configMock.On("MapKeyToString", shareConfigMap, SettingShareUseCache).Return("yes")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool).Return("cache")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool2).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCOW).Return("auto")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareAllocator).Return("highwater")
	configMock.On("MapKeyToInt", shareConfigMap, SettingShareSplitLevel).Return(0)
	configMock.On("MapKeyToUInt64", shareConfigMap, SettingShareFloor).Return(uint64(10000000))
	configMock.On("MapKeyToString", shareConfigMap, SettingShareIncludeDisks).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareExcludeDisks).Return("")

	disks := map[string]*Disk{}

	pools := map[string]*Pool{
		"cache": {Name: "cache", FSPath: "/mnt/cache"},
	}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}

	shares, err := handler.establishShares(disks, pools)
	require.NoError(t, err, "establishShares should not return an error")
	require.NotNil(t, shares, "shares map should not be nil")

	share, ok := shares["share1"]
	require.True(t, ok, "share1 should be present")
	assert.Equal(t, "share1", share.Name)
	assert.Equal(t, "yes", share.UseCache)
	require.NotNil(t, share.CachePool, "CachePool should not be nil")
	assert.Equal(t, "cache", share.CachePool.Name)
	assert.Nil(t, share.CachePool2, "CachePool2 should be nil")
	assert.Equal(t, "highwater", share.Allocator)
	assert.False(t, share.DisableCOW)
	assert.Equal(t, 0, share.SplitLevel)
	assert.Equal(t, uint64(10000000), share.SpaceFloor)
	assert.Equal(t, disks, share.GetIncludedDisks(), "included disks should match all disks")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Success_ShareIncludes verifies that a valid share
// configuration file is processed correctly, when share-specific includes are
// set.
func TestEstablishShares_Success_ShareIncludes(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)

	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "",
		SettingGlobalShareExcludes: "",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("")

	shareFilePath := filepath.Join(ConfigDirShares, "share1.cfg")
	shareConfigMap := map[string]string{}

	configMock.On("ReadGeneric", shareFilePath).Return(shareConfigMap, nil)
	configMock.On("MapKeyToString", shareConfigMap, SettingShareUseCache).Return("yes")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool).Return("cache")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool2).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCOW).Return("auto")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareAllocator).Return("highwater")
	configMock.On("MapKeyToInt", shareConfigMap, SettingShareSplitLevel).Return(0)
	configMock.On("MapKeyToUInt64", shareConfigMap, SettingShareFloor).Return(uint64(10000000))
	configMock.On("MapKeyToString", shareConfigMap, SettingShareIncludeDisks).Return("disk1,disk2")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareExcludeDisks).Return("")

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
		"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
		"disk3": {Name: "disk3", FSPath: "/mnt/disk3"},
		"disk4": {Name: "disk4", FSPath: "/mnt/disk4"},
	}

	expectedDisks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
		"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
	}

	pools := map[string]*Pool{
		"cache": {Name: "cache", FSPath: "/mnt/cache"},
	}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}

	shares, err := handler.establishShares(disks, pools)
	require.NoError(t, err, "establishShares should not return an error")
	require.NotNil(t, shares, "shares map should not be nil")

	share, ok := shares["share1"]
	require.True(t, ok, "share1 should be present")
	assert.Equal(t, "share1", share.Name)
	assert.Equal(t, "yes", share.UseCache)
	require.NotNil(t, share.CachePool, "CachePool should not be nil")
	assert.Equal(t, "cache", share.CachePool.Name)
	assert.Nil(t, share.CachePool2, "CachePool2 should be nil")
	assert.Equal(t, "highwater", share.Allocator)
	assert.False(t, share.DisableCOW)
	assert.Equal(t, 0, share.SplitLevel)
	assert.Equal(t, uint64(10000000), share.SpaceFloor)
	assert.Equal(t, expectedDisks, share.GetIncludedDisks(), "included disks should match expected disks")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Success_ShareExcludes verifies that a valid share
// configuration file is processed correctly, when share-specific excludes are
// set.
func TestEstablishShares_Success_ShareExcludes(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)

	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "",
		SettingGlobalShareExcludes: "",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("")

	shareFilePath := filepath.Join(ConfigDirShares, "share1.cfg")
	shareConfigMap := map[string]string{}

	configMock.On("ReadGeneric", shareFilePath).Return(shareConfigMap, nil)
	configMock.On("MapKeyToString", shareConfigMap, SettingShareUseCache).Return("yes")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool).Return("cache")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool2).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCOW).Return("auto")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareAllocator).Return("highwater")
	configMock.On("MapKeyToInt", shareConfigMap, SettingShareSplitLevel).Return(0)
	configMock.On("MapKeyToUInt64", shareConfigMap, SettingShareFloor).Return(uint64(10000000))
	configMock.On("MapKeyToString", shareConfigMap, SettingShareIncludeDisks).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareExcludeDisks).Return("disk3")

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
		"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
		"disk3": {Name: "disk3", FSPath: "/mnt/disk3"},
		"disk4": {Name: "disk4", FSPath: "/mnt/disk4"},
	}

	expectedDisks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
		"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
		"disk4": {Name: "disk4", FSPath: "/mnt/disk4"},
	}

	pools := map[string]*Pool{
		"cache": {Name: "cache", FSPath: "/mnt/cache"},
	}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}

	shares, err := handler.establishShares(disks, pools)
	require.NoError(t, err, "establishShares should not return an error")
	require.NotNil(t, shares, "shares map should not be nil")

	share, ok := shares["share1"]
	require.True(t, ok, "share1 should be present")
	assert.Equal(t, "share1", share.Name)
	assert.Equal(t, "yes", share.UseCache)
	require.NotNil(t, share.CachePool, "CachePool should not be nil")
	assert.Equal(t, "cache", share.CachePool.Name)
	assert.Nil(t, share.CachePool2, "CachePool2 should be nil")
	assert.Equal(t, "highwater", share.Allocator)
	assert.False(t, share.DisableCOW)
	assert.Equal(t, 0, share.SplitLevel)
	assert.Equal(t, uint64(10000000), share.SpaceFloor)
	assert.Equal(t, expectedDisks, share.GetIncludedDisks(), "included disks should match expected disks")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Success_GlobalIncludes verifies that a valid share
// configuration file is processed correctly, when global includes are set.
func TestEstablishShares_Success_GlobalIncludes(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)

	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "disk3",
		SettingGlobalShareExcludes: "",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("disk3")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("")

	shareFilePath := filepath.Join(ConfigDirShares, "share1.cfg")
	shareConfigMap := map[string]string{}

	configMock.On("ReadGeneric", shareFilePath).Return(shareConfigMap, nil)
	configMock.On("MapKeyToString", shareConfigMap, SettingShareUseCache).Return("yes")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool).Return("cache")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool2).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCOW).Return("auto")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareAllocator).Return("highwater")
	configMock.On("MapKeyToInt", shareConfigMap, SettingShareSplitLevel).Return(0)
	configMock.On("MapKeyToUInt64", shareConfigMap, SettingShareFloor).Return(uint64(10000000))
	configMock.On("MapKeyToString", shareConfigMap, SettingShareIncludeDisks).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareExcludeDisks).Return("")

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
		"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
		"disk3": {Name: "disk3", FSPath: "/mnt/disk3"},
		"disk4": {Name: "disk4", FSPath: "/mnt/disk4"},
	}

	expectedDisks := map[string]*Disk{
		"disk3": {Name: "disk3", FSPath: "/mnt/disk3"},
	}

	pools := map[string]*Pool{
		"cache": {Name: "cache", FSPath: "/mnt/cache"},
	}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}

	shares, err := handler.establishShares(disks, pools)
	require.NoError(t, err, "establishShares should not return an error")
	require.NotNil(t, shares, "shares map should not be nil")

	share, ok := shares["share1"]
	require.True(t, ok, "share1 should be present")
	assert.Equal(t, "share1", share.Name)
	assert.Equal(t, "yes", share.UseCache)
	require.NotNil(t, share.CachePool, "CachePool should not be nil")
	assert.Equal(t, "cache", share.CachePool.Name)
	assert.Nil(t, share.CachePool2, "CachePool2 should be nil")
	assert.Equal(t, "highwater", share.Allocator)
	assert.False(t, share.DisableCOW)
	assert.Equal(t, 0, share.SplitLevel)
	assert.Equal(t, uint64(10000000), share.SpaceFloor)
	assert.Equal(t, expectedDisks, share.GetIncludedDisks(), "included disks should match expected disks")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Success_GlobalExcludes verifies that a valid share
// configuration file is processed correctly, when global excludes are set.
func TestEstablishShares_Success_GlobalExcludes(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)

	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "",
		SettingGlobalShareExcludes: "disk3",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("disk3")

	shareFilePath := filepath.Join(ConfigDirShares, "share1.cfg")
	shareConfigMap := map[string]string{}

	configMock.On("ReadGeneric", shareFilePath).Return(shareConfigMap, nil)
	configMock.On("MapKeyToString", shareConfigMap, SettingShareUseCache).Return("yes")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool).Return("cache")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool2).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCOW).Return("auto")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareAllocator).Return("highwater")
	configMock.On("MapKeyToInt", shareConfigMap, SettingShareSplitLevel).Return(0)
	configMock.On("MapKeyToUInt64", shareConfigMap, SettingShareFloor).Return(uint64(10000000))
	configMock.On("MapKeyToString", shareConfigMap, SettingShareIncludeDisks).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareExcludeDisks).Return("")

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
		"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
		"disk3": {Name: "disk3", FSPath: "/mnt/disk3"},
		"disk4": {Name: "disk4", FSPath: "/mnt/disk4"},
	}

	expectedDisks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
		"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
		"disk4": {Name: "disk4", FSPath: "/mnt/disk4"},
	}

	pools := map[string]*Pool{
		"cache": {Name: "cache", FSPath: "/mnt/cache"},
	}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}

	shares, err := handler.establishShares(disks, pools)
	require.NoError(t, err, "establishShares should not return an error")
	require.NotNil(t, shares, "shares map should not be nil")

	share, ok := shares["share1"]
	require.True(t, ok, "share1 should be present")
	assert.Equal(t, "share1", share.Name)
	assert.Equal(t, "yes", share.UseCache)
	require.NotNil(t, share.CachePool, "CachePool should not be nil")
	assert.Equal(t, "cache", share.CachePool.Name)
	assert.Nil(t, share.CachePool2, "CachePool2 should be nil")
	assert.Equal(t, "highwater", share.Allocator)
	assert.False(t, share.DisableCOW)
	assert.Equal(t, 0, share.SplitLevel)
	assert.Equal(t, uint64(10000000), share.SpaceFloor)
	assert.Equal(t, expectedDisks, share.GetIncludedDisks(), "included disks should match expected disks")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Success_MixedIncludesExcludes_1 verifies that a valid share
// configuration file is processed correctly, when mixed includes/excludes are set.
func TestEstablishShares_Success_MixedIncludesExcludes_1(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)

	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "disk1,disk2,disk6",
		SettingGlobalShareExcludes: "disk3",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("disk1,disk2,disk6")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("disk3")

	shareFilePath := filepath.Join(ConfigDirShares, "share1.cfg")
	shareConfigMap := map[string]string{}

	configMock.On("ReadGeneric", shareFilePath).Return(shareConfigMap, nil)
	configMock.On("MapKeyToString", shareConfigMap, SettingShareUseCache).Return("yes")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool).Return("cache")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool2).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCOW).Return("auto")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareAllocator).Return("highwater")
	configMock.On("MapKeyToInt", shareConfigMap, SettingShareSplitLevel).Return(0)
	configMock.On("MapKeyToUInt64", shareConfigMap, SettingShareFloor).Return(uint64(10000000))
	configMock.On("MapKeyToString", shareConfigMap, SettingShareIncludeDisks).Return("disk1")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareExcludeDisks).Return("disk5,disk6")

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
		"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
		"disk3": {Name: "disk3", FSPath: "/mnt/disk3"},
		"disk4": {Name: "disk4", FSPath: "/mnt/disk4"},
		"disk5": {Name: "disk5", FSPath: "/mnt/disk5"},
		"disk6": {Name: "disk6", FSPath: "/mnt/disk6"},
	}

	expectedDisks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
	}

	pools := map[string]*Pool{
		"cache": {Name: "cache", FSPath: "/mnt/cache"},
	}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}

	shares, err := handler.establishShares(disks, pools)
	require.NoError(t, err, "establishShares should not return an error")
	require.NotNil(t, shares, "shares map should not be nil")

	share, ok := shares["share1"]
	require.True(t, ok, "share1 should be present")
	assert.Equal(t, "share1", share.Name)
	assert.Equal(t, "yes", share.UseCache)
	require.NotNil(t, share.CachePool, "CachePool should not be nil")
	assert.Equal(t, "cache", share.CachePool.Name)
	assert.Nil(t, share.CachePool2, "CachePool2 should be nil")
	assert.Equal(t, "highwater", share.Allocator)
	assert.False(t, share.DisableCOW)
	assert.Equal(t, 0, share.SplitLevel)
	assert.Equal(t, uint64(10000000), share.SpaceFloor)
	assert.Equal(t, expectedDisks, share.GetIncludedDisks(), "included disks should match expected disks")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Success_MixedIncludesExcludes_2 verifies that a valid share
// configuration file is processed correctly, when mixed includes/excludes are set.
func TestEstablishShares_Success_MixedIncludesExcludes_2(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)

	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "disk1,disk2,disk6",
		SettingGlobalShareExcludes: "disk3",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("disk1,disk2,disk6")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("disk3")

	shareFilePath := filepath.Join(ConfigDirShares, "share1.cfg")
	shareConfigMap := map[string]string{}

	configMock.On("ReadGeneric", shareFilePath).Return(shareConfigMap, nil)
	configMock.On("MapKeyToString", shareConfigMap, SettingShareUseCache).Return("yes")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool).Return("cache")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool2).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCOW).Return("auto")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareAllocator).Return("highwater")
	configMock.On("MapKeyToInt", shareConfigMap, SettingShareSplitLevel).Return(0)
	configMock.On("MapKeyToUInt64", shareConfigMap, SettingShareFloor).Return(uint64(10000000))
	configMock.On("MapKeyToString", shareConfigMap, SettingShareIncludeDisks).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareExcludeDisks).Return("disk5,disk6")

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
		"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
		"disk3": {Name: "disk3", FSPath: "/mnt/disk3"},
		"disk4": {Name: "disk4", FSPath: "/mnt/disk4"},
		"disk5": {Name: "disk5", FSPath: "/mnt/disk5"},
		"disk6": {Name: "disk6", FSPath: "/mnt/disk6"},
	}

	expectedDisks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
		"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
	}

	pools := map[string]*Pool{
		"cache": {Name: "cache", FSPath: "/mnt/cache"},
	}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}

	shares, err := handler.establishShares(disks, pools)
	require.NoError(t, err, "establishShares should not return an error")
	require.NotNil(t, shares, "shares map should not be nil")

	share, ok := shares["share1"]
	require.True(t, ok, "share1 should be present")
	assert.Equal(t, "share1", share.Name)
	assert.Equal(t, "yes", share.UseCache)
	require.NotNil(t, share.CachePool, "CachePool should not be nil")
	assert.Equal(t, "cache", share.CachePool.Name)
	assert.Nil(t, share.CachePool2, "CachePool2 should be nil")
	assert.Equal(t, "highwater", share.Allocator)
	assert.False(t, share.DisableCOW)
	assert.Equal(t, 0, share.SplitLevel)
	assert.Equal(t, uint64(10000000), share.SpaceFloor)
	assert.Equal(t, expectedDisks, share.GetIncludedDisks(), "included disks should match expected disks")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Fail_ConfigDirDoesNotExist simulates that the share
// configuration directory missing.
func TestEstablishShares_Fail_ConfigDirDoesNotExist(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(false, nil)

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}
	shares, err := handler.establishShares(nil, nil)
	require.Error(t, err, "an error is expected when config dir does not exist")
	assert.Nil(t, shares)
	assert.Contains(t, err.Error(), "config dir does not exist", "error should mention missing config dir")

	fsMock.AssertExpectations(t)
}

// TestEstablishShares_Fail_ReadDirError simulates an error while reading the share
// configuation directory.
func TestEstablishShares_Fail_ReadDirError(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	readErr := errors.New("read error")
	osMock.On("ReadDir", ConfigDirShares).Return(nil, readErr)

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}
	shares, err := handler.establishShares(nil, nil)
	require.Error(t, err, "an error is expected when ReadDir fails")
	assert.Nil(t, shares)
	assert.Contains(t, err.Error(), "failed to readdir", "error should mention readdir failure")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
}

// TestEstablishShares_Fail_GlobalConfigError simulates a failure reading the global
// share configuration.
func TestEstablishShares_Fail_GlobalConfigError(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(nil, errors.New("global read error"))

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}
	shares, err := handler.establishShares(nil, nil)
	require.Error(t, err, "an error is expected when global config read fails")
	assert.Nil(t, shares)
	assert.Contains(t, err.Error(), "failed to establish global share config", "error should mention global share config")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Fail_ShareConfigError simulates an error reading an individual
// share configuration file.
func TestEstablishShares_Fail_ShareConfigError(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)

	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "",
		SettingGlobalShareExcludes: "",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("")

	shareFilePath := filepath.Join(ConfigDirShares, "share1.cfg")
	configMock.On("ReadGeneric", shareFilePath).Return(nil, errors.New("share read error"))

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}
	shares, err := handler.establishShares(nil, nil)
	require.Error(t, err, "an error is expected when share config read fails")
	assert.Nil(t, shares)
	assert.Contains(t, err.Error(), "failed to read config", "error should mention share config read failure")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Fail_PrimaryCacheError simulates a failure to resolve the
// primary cache pool. Here we pass a pools map that does not contain the key
// returned by the share config.
func TestEstablishShares_Fail_PrimaryCacheError(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)
	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "",
		SettingGlobalShareExcludes: "",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("")

	shareFilePath := filepath.Join(ConfigDirShares, "share1.cfg")
	shareConfigMap := map[string]string{}

	configMock.On("ReadGeneric", shareFilePath).Return(shareConfigMap, nil)
	configMock.On("MapKeyToString", shareConfigMap, SettingShareUseCache).Return("yes")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool).Return("cache")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCOW).Return("auto")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareAllocator).Return("highwater")
	configMock.On("MapKeyToInt", shareConfigMap, SettingShareSplitLevel).Return(0)
	configMock.On("MapKeyToUInt64", shareConfigMap, SettingShareFloor).Return(uint64(10000000))

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
	}
	pools := map[string]*Pool{}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}
	shares, err := handler.establishShares(disks, pools)
	require.Error(t, err, "an error is expected when primary cache pool cannot be resolved")
	assert.Nil(t, shares)
	assert.Contains(t, err.Error(), "failed to deref primary cache for share", "error should mention primary cache resolution failure")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Fail_SecondaryCacheError simulates a failure to resolve the
// secondary cache pool. Here we pass a pools map that does not contain the key
// returned by the share config.
func TestEstablishShares_Fail_SecondaryCacheError(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)
	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "",
		SettingGlobalShareExcludes: "",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("")

	shareFilePath := filepath.Join(ConfigDirShares, "share1.cfg")
	shareConfigMap := map[string]string{}

	configMock.On("ReadGeneric", shareFilePath).Return(shareConfigMap, nil)
	configMock.On("MapKeyToString", shareConfigMap, SettingShareUseCache).Return("yes")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool).Return("cache")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool2).Return("cache2")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCOW).Return("auto")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareAllocator).Return("highwater")
	configMock.On("MapKeyToInt", shareConfigMap, SettingShareSplitLevel).Return(0)
	configMock.On("MapKeyToUInt64", shareConfigMap, SettingShareFloor).Return(uint64(10000000))

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
	}
	pools := map[string]*Pool{
		"cache": {Name: "cache", FSPath: "/mnt/cache"},
	}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}
	shares, err := handler.establishShares(disks, pools)
	require.Error(t, err, "an error is expected when secondary cache pool cannot be resolved")
	assert.Nil(t, shares)
	assert.Contains(t, err.Error(), "failed to deref secondary cache for share", "error should mention secondary cache resolution failure")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Fail_ShareIncludeError simulates a failure to resolve the
// share's included disks. Here we pass a disks map that does not contain the
// key returned by the share config.
func TestEstablishShares_Fail_ShareIncludeError(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)

	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "",
		SettingGlobalShareExcludes: "",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("")

	shareFilePath := filepath.Join(ConfigDirShares, "share1.cfg")
	shareConfigMap := map[string]string{}

	configMock.On("ReadGeneric", shareFilePath).Return(shareConfigMap, nil)
	configMock.On("MapKeyToString", shareConfigMap, SettingShareUseCache).Return("yes")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool).Return("cache")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool2).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCOW).Return("auto")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareAllocator).Return("highwater")
	configMock.On("MapKeyToInt", shareConfigMap, SettingShareSplitLevel).Return(0)
	configMock.On("MapKeyToUInt64", shareConfigMap, SettingShareFloor).Return(uint64(10000000))
	configMock.On("MapKeyToString", shareConfigMap, SettingShareIncludeDisks).Return("disk3")

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
	}

	pools := map[string]*Pool{
		"cache": {Name: "cache", FSPath: "/mnt/cache"},
	}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}

	shares, err := handler.establishShares(disks, pools)
	require.Error(t, err, "an error is expected when a share included disk cannot be resolved")
	assert.Nil(t, shares)
	assert.Contains(t, err.Error(), "failed to deref included disks for share", "error should mention share included disk resolution failure")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Fail_ShareExcludeError simulates a failure to resolve the
// share's excluded disks. Here we pass a disks map that does not contain the
// key returned by the share config.
func TestEstablishShares_Fail_ShareExcludeError(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)

	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "",
		SettingGlobalShareExcludes: "",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("")

	shareFilePath := filepath.Join(ConfigDirShares, "share1.cfg")
	shareConfigMap := map[string]string{}

	configMock.On("ReadGeneric", shareFilePath).Return(shareConfigMap, nil)
	configMock.On("MapKeyToString", shareConfigMap, SettingShareUseCache).Return("yes")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool).Return("cache")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCachePool2).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareCOW).Return("auto")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareAllocator).Return("highwater")
	configMock.On("MapKeyToInt", shareConfigMap, SettingShareSplitLevel).Return(0)
	configMock.On("MapKeyToUInt64", shareConfigMap, SettingShareFloor).Return(uint64(10000000))
	configMock.On("MapKeyToString", shareConfigMap, SettingShareIncludeDisks).Return("")
	configMock.On("MapKeyToString", shareConfigMap, SettingShareExcludeDisks).Return("disk3")

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
	}

	pools := map[string]*Pool{
		"cache": {Name: "cache", FSPath: "/mnt/cache"},
	}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}

	shares, err := handler.establishShares(disks, pools)
	require.Error(t, err, "an error is expected when a share excluded disk cannot be resolved")
	assert.Nil(t, shares)
	assert.Contains(t, err.Error(), "failed to deref excluded disks for share", "error should mention share excluded disk resolution failure")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Fail_GlobalIncludeError simulates a failure to resolve the
// global included disks. Here we pass a disks map that does not contain the key
// returned by the share config.
func TestEstablishShares_Fail_GlobalIncludeError(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)

	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "disk3",
		SettingGlobalShareExcludes: "",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("disk3")

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
	}

	pools := map[string]*Pool{
		"cache": {Name: "cache", FSPath: "/mnt/cache"},
	}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}

	shares, err := handler.establishShares(disks, pools)
	require.Error(t, err, "an error is expected when a global included disk cannot be resolved")
	assert.Nil(t, shares)
	assert.Contains(t, err.Error(), "failed to deref global included disks", "error should mention global included disk resolution failure")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShares_Fail_GlobalExcludeError simulates a failure to resolve the
// global excluded disks. Here we pass a disks map that does not contain the key
// returned by the share config.
func TestEstablishShares_Fail_GlobalExcludeError(t *testing.T) {
	t.Parallel()

	fsMock := mocks.NewFsProvider(t)
	osMock := mocks.NewOsProvider(t)
	configMock := mocks.NewConfigProvider(t)

	fsMock.On("Exists", ConfigDirShares).Return(true, nil)
	fakeShareFile := fakeDirEntry{name: "share1.cfg", isDir: false}
	osMock.On("ReadDir", ConfigDirShares).Return([]os.DirEntry{fakeShareFile}, nil)

	globalConfigMap := map[string]string{
		SettingGlobalShareIncludes: "",
		SettingGlobalShareExcludes: "disk3",
	}
	configMock.On("ReadGeneric", GlobalShareConfigFile).Return(globalConfigMap, nil)
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareIncludes).Return("")
	configMock.On("MapKeyToString", globalConfigMap, SettingGlobalShareExcludes).Return("disk3")

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
	}

	pools := map[string]*Pool{
		"cache": {Name: "cache", FSPath: "/mnt/cache"},
	}

	handler := &Handler{
		fsHandler:     fsMock,
		osHandler:     osMock,
		configHandler: configMock,
	}

	shares, err := handler.establishShares(disks, pools)
	require.Error(t, err, "an error is expected when a global excluded disk cannot be resolved")
	assert.Nil(t, shares)
	assert.Contains(t, err.Error(), "failed to deref global excluded disks", "error should mention global excluded disk resolution failure")

	fsMock.AssertExpectations(t)
	osMock.AssertExpectations(t)
	configMock.AssertExpectations(t)
}

// TestEstablishShareIncludes tests the establishShareIncludes method.
func TestEstablishShareIncludes(t *testing.T) {
	t.Parallel()

	disk1 := &Disk{Name: "disk1", FSPath: "/mnt/disk1"}
	disk2 := &Disk{Name: "disk2", FSPath: "/mnt/disk2"}
	disk3 := &Disk{Name: "disk3", FSPath: "/mnt/disk3"}
	disk4 := &Disk{Name: "disk4", FSPath: "/mnt/disk4"}

	allDisks := map[string]*Disk{
		"disk1": disk1,
		"disk2": disk2,
		"disk3": disk3,
		"disk4": disk4,
	}

	tests := []struct {
		name           string
		allDisks       map[string]*Disk
		config         *includesExcludesConfig
		expectedResult map[string]*Disk
	}{
		{
			name:     "Success_AllDisksIncluded",
			allDisks: allDisks,
			config: &includesExcludesConfig{
				shareIncludes:  nil,
				shareExcludes:  map[string]*Disk{},
				globalIncludes: nil,
				globalExcludes: map[string]*Disk{},
			},
			expectedResult: map[string]*Disk{
				"disk1": disk1,
				"disk2": disk2,
				"disk3": disk3,
				"disk4": disk4,
			},
		},
		{
			name:     "Success_ShareIncludesOnly",
			allDisks: allDisks,
			config: &includesExcludesConfig{
				shareIncludes: map[string]*Disk{
					"disk1": disk1,
					"disk2": disk2,
				},
				shareExcludes:  map[string]*Disk{},
				globalIncludes: nil,
				globalExcludes: map[string]*Disk{},
			},
			expectedResult: map[string]*Disk{
				"disk1": disk1,
				"disk2": disk2,
			},
		},
		{
			name:     "Success_GlobalIncludesOnly",
			allDisks: allDisks,
			config: &includesExcludesConfig{
				shareIncludes: nil,
				shareExcludes: map[string]*Disk{},
				globalIncludes: map[string]*Disk{
					"disk1": disk1,
					"disk3": disk3,
				},
				globalExcludes: map[string]*Disk{},
			},
			expectedResult: map[string]*Disk{
				"disk1": disk1,
				"disk3": disk3,
			},
		},
		{
			name:     "Success_ShareExcludesOnly",
			allDisks: allDisks,
			config: &includesExcludesConfig{
				shareIncludes: nil,
				shareExcludes: map[string]*Disk{
					"disk3": disk3,
				},
				globalIncludes: nil,
				globalExcludes: map[string]*Disk{},
			},
			expectedResult: map[string]*Disk{
				"disk1": disk1,
				"disk2": disk2,
				"disk4": disk4,
			},
		},
		{
			name:     "Success_GlobalExcludesOnly",
			allDisks: allDisks,
			config: &includesExcludesConfig{
				shareIncludes:  nil,
				shareExcludes:  map[string]*Disk{},
				globalIncludes: nil,
				globalExcludes: map[string]*Disk{
					"disk4": disk4,
				},
			},
			expectedResult: map[string]*Disk{
				"disk1": disk1,
				"disk2": disk2,
				"disk3": disk3,
			},
		},
		{
			name:     "Success_ShareAndGlobalIncludes",
			allDisks: allDisks,
			config: &includesExcludesConfig{
				shareIncludes: map[string]*Disk{
					"disk1": disk1,
					"disk2": disk2,
					"disk3": disk3,
				},
				shareExcludes: map[string]*Disk{},
				globalIncludes: map[string]*Disk{
					"disk1": disk1,
					"disk3": disk3,
				},
				globalExcludes: map[string]*Disk{},
			},
			expectedResult: map[string]*Disk{
				"disk1": disk1,
				"disk3": disk3,
			},
		},
		{
			name:     "Success_ShareAndGlobalExcludes",
			allDisks: allDisks,
			config: &includesExcludesConfig{
				shareIncludes: nil,
				shareExcludes: map[string]*Disk{
					"disk2": disk2,
				},
				globalIncludes: nil,
				globalExcludes: map[string]*Disk{
					"disk4": disk4,
				},
			},
			expectedResult: map[string]*Disk{
				"disk1": disk1,
				"disk3": disk3,
			},
		},
		{
			name:     "Success_ShareIncludesAndExcludes",
			allDisks: allDisks,
			config: &includesExcludesConfig{
				shareIncludes: map[string]*Disk{
					"disk1": disk1,
					"disk2": disk2,
					"disk3": disk3,
				},
				shareExcludes: map[string]*Disk{
					"disk3": disk3,
				},
				globalIncludes: nil,
				globalExcludes: map[string]*Disk{},
			},
			expectedResult: map[string]*Disk{
				"disk1": disk1,
				"disk2": disk2,
			},
		},
		{
			name:     "Success_GlobalIncludesAndExcludes",
			allDisks: allDisks,
			config: &includesExcludesConfig{
				shareIncludes: nil,
				shareExcludes: map[string]*Disk{},
				globalIncludes: map[string]*Disk{
					"disk1": disk1,
					"disk2": disk2,
					"disk3": disk3,
				},
				globalExcludes: map[string]*Disk{
					"disk2": disk2,
				},
			},
			expectedResult: map[string]*Disk{
				"disk1": disk1,
				"disk3": disk3,
			},
		},
		{
			name:     "Success_ComplexCombination",
			allDisks: allDisks,
			config: &includesExcludesConfig{
				shareIncludes: map[string]*Disk{
					"disk1": disk1,
					"disk2": disk2,
					"disk3": disk3,
				},
				shareExcludes: map[string]*Disk{
					"disk3": disk3,
				},
				globalIncludes: map[string]*Disk{
					"disk1": disk1,
					"disk2": disk2,
				},
				globalExcludes: map[string]*Disk{
					"disk2": disk2,
				},
			},
			expectedResult: map[string]*Disk{
				"disk1": disk1,
			},
		},
		{
			name:     "Success_EmptyResult",
			allDisks: allDisks,
			config: &includesExcludesConfig{
				shareIncludes: map[string]*Disk{
					"disk1": disk1,
				},
				shareExcludes: map[string]*Disk{
					"disk1": disk1,
				},
				globalIncludes: nil,
				globalExcludes: map[string]*Disk{},
			},
			expectedResult: map[string]*Disk{},
		},
		{
			name:     "Success_EmptyAllDisks",
			allDisks: map[string]*Disk{},
			config: &includesExcludesConfig{
				shareIncludes:  nil,
				shareExcludes:  map[string]*Disk{},
				globalIncludes: nil,
				globalExcludes: map[string]*Disk{},
			},
			expectedResult: map[string]*Disk{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &Handler{}
			result := handler.establishShareIncludes(tt.allDisks, tt.config)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestEstablishGlobalIncludesExcludes tests the establishGlobalIncludesExcludes method.
func TestEstablishGlobalIncludesExcludes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		configMap           map[string]string
		configReadError     error
		disks               map[string]*Disk
		globalIncludesValue string
		globalExcludesValue string
		expectError         bool
		expectErrorContains string
		expectedGlobalInc   map[string]*Disk
		expectedGlobalExc   map[string]*Disk
	}{
		{
			name: "Success_EmptyValues",
			configMap: map[string]string{
				SettingGlobalShareIncludes: "",
				SettingGlobalShareExcludes: "",
			},
			configReadError: nil,
			disks: map[string]*Disk{
				"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
				"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
			},
			globalIncludesValue: "",
			globalExcludesValue: "",
			expectError:         false,
			expectedGlobalInc:   nil,
			expectedGlobalExc:   nil,
		},
		{
			name: "Success_OnlyIncludes",
			configMap: map[string]string{
				SettingGlobalShareIncludes: "disk1",
				SettingGlobalShareExcludes: "",
			},
			configReadError: nil,
			disks: map[string]*Disk{
				"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
				"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
			},
			globalIncludesValue: "disk1",
			globalExcludesValue: "",
			expectError:         false,
			expectedGlobalInc: map[string]*Disk{
				"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
			},
			expectedGlobalExc: nil,
		},
		{
			name: "Success_OnlyExcludes",
			configMap: map[string]string{
				SettingGlobalShareIncludes: "",
				SettingGlobalShareExcludes: "disk2",
			},
			configReadError: nil,
			disks: map[string]*Disk{
				"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
				"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
			},
			globalIncludesValue: "",
			globalExcludesValue: "disk2",
			expectError:         false,
			expectedGlobalInc:   nil,
			expectedGlobalExc: map[string]*Disk{
				"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
			},
		},
		{
			name: "Success_BothIncludesAndExcludes",
			configMap: map[string]string{
				SettingGlobalShareIncludes: "disk1",
				SettingGlobalShareExcludes: "disk2",
			},
			configReadError: nil,
			disks: map[string]*Disk{
				"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
				"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
			},
			globalIncludesValue: "disk1",
			globalExcludesValue: "disk2",
			expectError:         false,
			expectedGlobalInc: map[string]*Disk{
				"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
			},
			expectedGlobalExc: map[string]*Disk{
				"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
			},
		},
		{
			name: "Success_MultipleValues",
			configMap: map[string]string{
				SettingGlobalShareIncludes: "disk1,disk3",
				SettingGlobalShareExcludes: "disk2,disk4",
			},
			configReadError: nil,
			disks: map[string]*Disk{
				"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
				"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
				"disk3": {Name: "disk3", FSPath: "/mnt/disk3"},
				"disk4": {Name: "disk4", FSPath: "/mnt/disk4"},
			},
			globalIncludesValue: "disk1,disk3",
			globalExcludesValue: "disk2,disk4",
			expectError:         false,
			expectedGlobalInc: map[string]*Disk{
				"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
				"disk3": {Name: "disk3", FSPath: "/mnt/disk3"},
			},
			expectedGlobalExc: map[string]*Disk{
				"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
				"disk4": {Name: "disk4", FSPath: "/mnt/disk4"},
			},
		},
		{
			name:            "Fail_ConfigReadFails",
			configMap:       nil,
			configReadError: errors.New("config read failed"),
			disks: map[string]*Disk{
				"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
				"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
			},
			expectError:         true,
			expectErrorContains: "failed to read global share config",
		},
		{
			name: "Fail_InvalidIncludeDisk",
			configMap: map[string]string{
				SettingGlobalShareIncludes: "nonexistentdisk",
				SettingGlobalShareExcludes: "",
			},
			configReadError: nil,
			disks: map[string]*Disk{
				"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
				"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
			},
			globalIncludesValue: "nonexistentdisk",
			globalExcludesValue: "",
			expectError:         true,
			expectErrorContains: "failed to deref global included disks",
		},
		{
			name: "Fail_InvalidExcludeDisk",
			configMap: map[string]string{
				SettingGlobalShareIncludes: "",
				SettingGlobalShareExcludes: "nonexistentdisk",
			},
			configReadError: nil,
			disks: map[string]*Disk{
				"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
				"disk2": {Name: "disk2", FSPath: "/mnt/disk2"},
			},
			globalIncludesValue: "",
			globalExcludesValue: "nonexistentdisk",
			expectError:         true,
			expectErrorContains: "failed to deref global excluded disks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configMock := mocks.NewConfigProvider(t)

			configMock.On("ReadGeneric", GlobalShareConfigFile).Return(tt.configMap, tt.configReadError)

			if !tt.expectError || tt.expectErrorContains != "failed to read global share config" {
				configMock.On("MapKeyToString", tt.configMap, SettingGlobalShareIncludes).Return(tt.globalIncludesValue)
				if tt.expectErrorContains != "failed to deref global included disks" {
					configMock.On("MapKeyToString", tt.configMap, SettingGlobalShareExcludes).Return(tt.globalExcludesValue)
				}
			}

			handler := &Handler{
				configHandler: configMock,
			}

			result, err := handler.establishGlobalIncludesExcludes(tt.disks)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErrorContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedGlobalInc, result.globalIncludes)
				assert.Equal(t, tt.expectedGlobalExc, result.globalExcludes)
			}

			configMock.AssertExpectations(t)
		})
	}
}
