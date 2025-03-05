package io

func MergeProgressReports(target, source *ProgressReport) {
	if target == nil || source == nil {
		return
	}

	target.AnyProcessed = append(target.AnyProcessed, source.AnyProcessed...)
	target.DirsProcessed = append(target.DirsProcessed, source.DirsProcessed...)
	target.HardlinksProcessed = append(target.HardlinksProcessed, source.HardlinksProcessed...)
	target.MoveablesProcessed = append(target.MoveablesProcessed, source.MoveablesProcessed...)
	target.SymlinksProcessed = append(target.SymlinksProcessed, source.SymlinksProcessed...)
}
