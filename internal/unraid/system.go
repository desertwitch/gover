package unraid

import (
	"fmt"
	"maps"
)

// System is the top-level representation of an Unraid system.
type System struct {
	Array  *Array
	Pools  map[string]*Pool
	Shares map[string]*Share
}

// GetPools returns a copy of the map (map[poolName]*Pool) with all [Pool].
func (s *System) GetPools() map[string]*Pool {
	if s.Pools == nil {
		return nil
	}

	pools := make(map[string]*Pool)
	maps.Copy(pools, s.Pools)

	return pools
}

// GetShares returns a copy of the map (map[shareName]*Share) with all [Share].
func (s *System) GetShares() map[string]*Share {
	if s.Shares == nil {
		return nil
	}

	shares := make(map[string]*Share)
	maps.Copy(shares, s.Shares)

	return shares
}

// EstablishSystem returns a pointer to an established Unraid [System]. It is
// the principal method to retrieve all information from the Unraid system.
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
