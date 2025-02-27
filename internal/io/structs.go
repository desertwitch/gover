package io

import "github.com/desertwitch/gover/internal/filesystem"

type InternalProgressReport struct {
	AnyProcessed       []filesystem.RelatedElement
	DirsProcessed      []*filesystem.RelatedDirectory
	MoveablesProcessed []*filesystem.Moveable
	SymlinksProcessed  []*filesystem.Moveable
	HardlinksProcessed []*filesystem.Moveable
}
