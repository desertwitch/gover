package unraid

import (
	"errors"
	"testing"

	"github.com/desertwitch/gover/internal/unraid/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEstablishArray_Success simulates success in establishing the array.
func TestEstablishArray_Success(t *testing.T) {
	t.Parallel()

	configMock := mocks.NewConfigProvider(t)

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
	}

	handler := &Handler{
		configHandler: configMock,
	}

	configMap := map[string]string{}

	configMock.On("ReadGeneric", ArrayStateFile).Return(configMap, nil)
	configMock.On("MapKeyToString", configMap, StateArrayStatus).Return("STARTED")
	configMock.On("MapKeyToString", configMap, StateTurboSetting).Return("auto")
	configMock.On("MapKeyToInt", configMap, StateParityPosition).Return(1)

	array, err := handler.establishArray(disks)
	require.NoError(t, err, "establishArray should not return an error")
	require.NotNil(t, array, "returned array should not be nil")

	assert.Equal(t, disks, array.Disks, "disks map mismatch")
	assert.Equal(t, "STARTED", array.Status, "unexpected array status")
	assert.Equal(t, "auto", array.TurboSetting, "unexpected turbo setting")
	assert.True(t, array.ParityRunning, "expected parity running to be true")

	configMock.AssertExpectations(t)
}

// TestEstablishArray_ReadConfigError simulates an error reading the array
// state file.
func TestEstablishArray_Fail_ReadConfigError(t *testing.T) {
	t.Parallel()

	configMock := mocks.NewConfigProvider(t)
	handler := &Handler{
		configHandler: configMock,
	}

	disks := map[string]*Disk{
		"disk1": {Name: "disk1", FSPath: "/mnt/disk1"},
	}

	readErr := errors.New("read error")
	configMock.On("ReadGeneric", ArrayStateFile).Return(nil, readErr)

	array, err := handler.establishArray(disks)
	require.Error(t, err, "an error was expected")
	assert.Nil(t, array, "array should be nil on error")
	assert.Contains(t, err.Error(), "failed to load array state file", "error message does not match expected output")

	configMock.AssertExpectations(t)
}
