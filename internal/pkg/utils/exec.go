package utils

import (
	"os/exec"
)

func Exec(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
