package utils

import (
	"os/exec"
	"path/filepath"
)

func LookupExecutablePathAbs(executable string) (string, error) {
	file, err := exec.LookPath(executable)
	if err != nil {
		return "", err
	}

	return filepath.Abs(file)
}
