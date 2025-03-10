package io

import "github.com/desertwitch/gover/internal/filesystem"

func mergeProgressReports(target, source *ProgressReport) {
	if target == nil || source == nil {
		return
	}

	target.AnyCreated = append(target.AnyCreated, source.AnyCreated...)
	target.DirsCreated = append(target.DirsCreated, source.DirsCreated...)
	target.DirsProcessed = append(target.DirsProcessed, source.DirsProcessed...)
	target.HardlinksCreated = append(target.HardlinksCreated, source.HardlinksCreated...)
	target.MoveablesCreated = append(target.MoveablesCreated, source.MoveablesCreated...)
	target.SymlinksCreated = append(target.SymlinksCreated, source.SymlinksCreated...)
}

func addToProgressReport(p *ProgressReport, m *filesystem.Moveable) {
	switch {
	case m.IsHardlink:
		p.AnyCreated = append(p.AnyCreated, m)
		p.HardlinksCreated = append(p.HardlinksCreated, m)

	case m.IsSymlink || m.Metadata.IsSymlink:
		p.AnyCreated = append(p.AnyCreated, m)
		p.SymlinksCreated = append(p.SymlinksCreated, m)

	default:
		p.AnyCreated = append(p.AnyCreated, m)
		p.MoveablesCreated = append(p.MoveablesCreated, m)
	}
}
