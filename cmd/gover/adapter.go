package main

import (
	"github.com/desertwitch/gover/internal/generic/storage"
	"github.com/desertwitch/gover/internal/unraid"
)

type ShareAdapter struct {
	*unraid.Share
}

func NewShareAdapter(original *unraid.Share) *ShareAdapter {
	return &ShareAdapter{Share: original}
}

func (s *ShareAdapter) GetCachePool() storage.Pool {
	if s.CachePool == nil {
		return nil
	}

	return s.CachePool
}

func (s *ShareAdapter) GetCachePool2() storage.Pool {
	if s.CachePool2 == nil {
		return nil
	}

	return s.CachePool2
}

func (a *ShareAdapter) GetIncludedDisks() map[string]storage.Disk {
	if a.Share.IncludedDisks == nil {
		return nil
	}

	result := make(map[string]storage.Disk, len(a.Share.IncludedDisks))
	for k, v := range a.Share.IncludedDisks {
		if v == nil {
			result[k] = nil

			continue
		}
		result[k] = v
	}

	return result
}

func (a *ShareAdapter) GetExcludedDisks() map[string]storage.Disk {
	if a.Share.ExcludedDisks == nil {
		return nil
	}

	result := make(map[string]storage.Disk, len(a.Share.ExcludedDisks))
	for k, v := range a.Share.ExcludedDisks {
		if v == nil {
			result[k] = nil

			continue
		}
		result[k] = v
	}

	return result
}
