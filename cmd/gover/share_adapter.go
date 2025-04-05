package main

import (
	"github.com/desertwitch/gover/internal/schema"
	"github.com/desertwitch/gover/internal/unraid"
)

type shareAdapter struct {
	*unraid.Share
}

func newShareAdapter(unraidShare *unraid.Share) *shareAdapter {
	return &shareAdapter{unraidShare}
}

func (s *shareAdapter) GetCachePool() schema.Pool { //nolint:ireturn
	if s.CachePool == nil {
		return nil
	}

	return s.CachePool
}

func (s *shareAdapter) GetCachePool2() schema.Pool { //nolint:ireturn
	if s.CachePool2 == nil {
		return nil
	}

	return s.CachePool2
}

// GetIncludedDisks returns a copy of the map containing disks that implement
// the schema.Disk interface.
func (s *shareAdapter) GetIncludedDisks() map[string]schema.Disk {
	if s.Share.IncludedDisks == nil {
		return nil
	}

	result := make(map[string]schema.Disk, len(s.Share.IncludedDisks))
	for k, v := range s.Share.IncludedDisks {
		if v == nil {
			result[k] = nil

			continue
		}
		result[k] = v
	}

	return result
}
