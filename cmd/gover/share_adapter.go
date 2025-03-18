package main

import (
	"github.com/desertwitch/gover/internal/generic/schema"
	"github.com/desertwitch/gover/internal/unraid"
)

type ShareAdapter struct {
	*unraid.Share
}

func NewShareAdapter(unraidShare *unraid.Share) *ShareAdapter {
	return &ShareAdapter{unraidShare}
}

func (s *ShareAdapter) GetCachePool() schema.Pool {
	if s.CachePool == nil {
		return nil
	}

	return s.CachePool
}

func (s *ShareAdapter) GetCachePool2() schema.Pool {
	if s.CachePool2 == nil {
		return nil
	}

	return s.CachePool2
}

// GetIncludedDisks returns a copy of the map containing disks that implement the schema.Disk interface.
func (s *ShareAdapter) GetIncludedDisks() map[string]schema.Disk {
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
