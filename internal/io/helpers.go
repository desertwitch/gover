package io

import "os/exec"

func (i *IOHandler) IsFileInUse(path string) (bool, error) {
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
