package io

import "github.com/desertwitch/gover/internal/filesystem"

func mergeProgressReports(target, source *ProgressReport) {
	if target == nil || source == nil {
		return
	}

	target.AnyProcessed = append(target.AnyProcessed, source.AnyProcessed...)
	target.DirsProcessed = append(target.DirsProcessed, source.DirsProcessed...)
	target.HardlinksProcessed = append(target.HardlinksProcessed, source.HardlinksProcessed...)
	target.MoveablesProcessed = append(target.MoveablesProcessed, source.MoveablesProcessed...)
	target.SymlinksProcessed = append(target.SymlinksProcessed, source.SymlinksProcessed...)
}

func addToProgressReport(p *ProgressReport, m *filesystem.Moveable) {
	if m.IsHardlink {
		p.AnyProcessed = append(p.AnyProcessed, m)
		p.HardlinksProcessed = append(p.HardlinksProcessed, m)
	} else if m.IsSymlink || m.Metadata.IsSymlink {
		p.AnyProcessed = append(p.AnyProcessed, m)
		p.SymlinksProcessed = append(p.SymlinksProcessed, m)
	} else {
		p.AnyProcessed = append(p.AnyProcessed, m)
		p.MoveablesProcessed = append(p.MoveablesProcessed, m)
	}
}
