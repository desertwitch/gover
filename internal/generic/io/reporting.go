package io

import (
	"github.com/desertwitch/gover/internal/generic/schema"
)

type ioReport struct {
	AnyCreated       []fsElement
	DirsCreated      []*schema.Directory
	DirsWalked       []*schema.Directory
	MoveablesCreated []*schema.Moveable
	SymlinksCreated  []*schema.Moveable
	HardlinksCreated []*schema.Moveable
}

func mergeIOReports(target, source *ioReport) {
	if target == nil || source == nil {
		return
	}

	target.AnyCreated = append(target.AnyCreated, source.AnyCreated...)
	target.DirsCreated = append(target.DirsCreated, source.DirsCreated...)
	target.DirsWalked = append(target.DirsWalked, source.DirsWalked...)
	target.HardlinksCreated = append(target.HardlinksCreated, source.HardlinksCreated...)
	target.MoveablesCreated = append(target.MoveablesCreated, source.MoveablesCreated...)
	target.SymlinksCreated = append(target.SymlinksCreated, source.SymlinksCreated...)
}

func addToIOReport(r *ioReport, m *schema.Moveable) {
	switch {
	case m.IsHardlink:
		r.AnyCreated = append(r.AnyCreated, m)
		r.HardlinksCreated = append(r.HardlinksCreated, m)

	case m.IsSymlink || m.Metadata.IsSymlink:
		r.AnyCreated = append(r.AnyCreated, m)
		r.SymlinksCreated = append(r.SymlinksCreated, m)

	default:
		r.AnyCreated = append(r.AnyCreated, m)
		r.MoveablesCreated = append(r.MoveablesCreated, m)
	}
}
