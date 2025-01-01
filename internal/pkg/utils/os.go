package utils

import (
	"os/user"
	"path/filepath"
)

func GetHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	homeDir := usr.HomeDir

	absPath, err := filepath.Abs(homeDir)
	if err != nil {
		return "", err
	}

	return absPath, nil
}
