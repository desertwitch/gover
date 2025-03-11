package unraid

import (
	"fmt"

	"github.com/desertwitch/gover/internal/filesystem"
)

type System struct {
	Array  *Array
	Pools  map[string]*Pool
	Shares map[string]*Share
}

func (s *System) GetArray() filesystem.ArrayType {
	if s.Array == nil {
		return nil
	}

	return s.Array
}

func (s *System) GetShares() map[string]filesystem.ShareType {
	if s.Shares == nil {
		return nil
	}

	shares := make(map[string]filesystem.ShareType)
	for k, v := range s.Shares {
		if v != nil {
			shares[k] = v
		}
	}

	return shares
}

// establishSystem returns a pointer to an established Unraid system.
func (u *Handler) EstablishSystem() (*System, error) {
	disks, err := u.establishDisks()
	if err != nil {
		return nil, fmt.Errorf("(unraid) failed establishing disks: %w", err)
	}

	pools, err := u.establishPools()
	if err != nil {
		return nil, fmt.Errorf("(unraid) failed establishing pools: %w", err)
	}

	shares, err := u.establishShares(disks, pools)
	if err != nil {
		return nil, fmt.Errorf("(unraid) failed establishing shares: %w", err)
	}

	array, err := u.establishArray(disks)
	if err != nil {
		return nil, fmt.Errorf("(unraid) failed establishing array: %w", err)
	}

	system := &System{
		Array:  array,
		Pools:  pools,
		Shares: shares,
	}

	return system, nil
}
