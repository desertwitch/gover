package unraid

import (
	"fmt"
)

type System struct {
	Array  *Array
	Pools  map[string]*Pool
	Shares map[string]*Share
}

// GetPools returns a copy of the pool map holding pointers to all pools.
func (s *System) GetPools() map[string]*Pool {
	if s.Pools == nil {
		return nil
	}

	pools := make(map[string]*Pool)

	for k, v := range s.Pools {
		pools[k] = v
	}

	return pools
}

// GetShares returns a copy of the share map holding pointers to all shares.
func (s *System) GetShares() map[string]*Share {
	if s.Shares == nil {
		return nil
	}

	shares := make(map[string]*Share)

	for k, v := range s.Shares {
		shares[k] = v
	}

	return shares
}

// establishSystem returns a pointer to an established Unraid system
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
