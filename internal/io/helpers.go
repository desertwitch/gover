package io

import (
	"os/exec"

	"github.com/desertwitch/gover/internal/filesystem"
)

func isFileInUse(path string) (bool, error) {
	cmd := exec.Command("lsof", path)

	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return false, nil
	}

	return false, err
}

func calculateDirectoryDepth(dir *filesystem.RelatedDirectory) int {
	depth := 0
	for dir != nil {
		dir = dir.Parent
		depth++
	}
	return depth
}
