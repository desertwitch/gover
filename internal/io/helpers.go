package io

import (
	"errors"
	"os/exec"
)

func (i *Handler) IsFileInUse(path string) (bool, error) {
	cmd := exec.Command("lsof", path)

	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}

	return false, err
}

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
