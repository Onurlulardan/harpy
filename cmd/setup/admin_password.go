package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func prepareAdminPassword(root, profile string, rotate bool) (string, error) {
	existing := ""
	if !rotate {
		stored, err := readAdminPassword(root, profile)
		if err != nil {
			return "", err
		}
		existing = stored
	}
	return reuseOrGenerate(existing, rotate, secretAlphabet, passwordLength)
}

func readAdminPassword(root, profile string) (string, error) {
	path := filepath.Join(root, "deploy", profile, "secrets", "initial_admin_password")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\r\n"), nil
}
