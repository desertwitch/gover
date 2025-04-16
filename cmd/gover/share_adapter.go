package main

import (
	"github.com/desertwitch/gover/internal/processors"
	"github.com/desertwitch/gover/internal/schema"
	"github.com/desertwitch/gover/internal/unraid"
)

// shareAdapter is an adapter for translating an [unraid.Share], with its
// internal structures, into [schema]-compatible implementations to use with
// more generic storage-system interfaces.
type shareAdapter struct {
	*unraid.Share
}

// newShareAdapter returns a pointer to a new [shareAdapter] for the
// [unraid.Share].
func newShareAdapter(unraidShare *unraid.Share) *shareAdapter {
	return &shareAdapter{unraidShare}
}

// GetCachePool returns the primary cache [unraid.Pool] implementing
// [schema.Pool]. If that [unraid.Pool] is nil, the primary cache is the
// [unraid.Array] (just for interpretation, the pointer is nil).
func (s *shareAdapter) GetCachePool() schema.Pool { //nolint:ireturn
	if s.CachePool == nil {
		return nil
	}

	return s.CachePool
}

// GetCachePool2 returns the secondary cache [unraid.Pool] implementing
// [schema.Pool]. If that [unraid.Pool] is nil, the secondary cache is the
// [unraid.Array] (just for interpretation, the pointer is nil).
func (s *shareAdapter) GetCachePool2() schema.Pool { //nolint:ireturn
	if s.CachePool2 == nil {
		return nil
	}

	return s.CachePool2
}

// GetIncludedDisks returns a map (map[diskName]schema.Disk) that is a copy of
// the internal map (map[diskName]unraid.Disk), but implementing [schema.Disk].
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

// GetPipeline returns the [schema.Pipeline] for the share.
func (s *shareAdapter) GetPipeline() schema.Pipeline { //nolint:ireturn
	pipeline := &processors.Pipeline{}

	return pipeline
}
