package unraid

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetPools tests the [System.GetPools] method.
func TestGetPools(t *testing.T) {
	t.Parallel()

	// When Pools is nil, GetPools should return nil.
	sys := &System{Pools: nil}
	assert.Nil(t, sys.GetPools(), "Expected nil when Pools is nil")

	// When Pools is set, GetPools should return a copy.
	origPools := map[string]*Pool{
		"pool1": {Name: "pool1", FSPath: "/mnt/pool1"},
		"pool2": {Name: "pool2", FSPath: "/mnt/pool2"},
	}
	sys = &System{Pools: origPools}
	copiedPools := sys.GetPools()
	assert.Equal(t, origPools, copiedPools, "Copied pools should equal original")
	assert.NotEqual(t, fmt.Sprintf("%p", origPools), fmt.Sprintf("%p", copiedPools), "Copied pools map should not be the same instance")
}

// TestGetShares tests the [System.GetShares] method.
func TestGetShares(t *testing.T) {
	t.Parallel()

	// When Shares is nil, GetShares should return nil.
	sys := &System{Shares: nil}
	assert.Nil(t, sys.GetShares(), "Expected nil when Shares is nil")

	// When Shares is set, GetShares should return a copy.
	origShares := map[string]*Share{
		"share1": {Name: "share1", UseCache: "yes"},
		"share2": {Name: "share2", UseCache: "no"},
	}
	sys = &System{Shares: origShares}
	copiedShares := sys.GetShares()
	assert.Equal(t, origShares, copiedShares, "Copied shares should equal original")
	assert.NotEqual(t, fmt.Sprintf("%p", origShares), fmt.Sprintf("%p", copiedShares), "Copied pools map should not be the same instance")
}
