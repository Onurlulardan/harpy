package main

import (
	"errors"
	"os"
	"path/filepath"
)

// setupRotateSecrets is set by run for one setup invocation.
// Direct runtime and release calls keep installed files.
var setupRotateSecrets bool

func writeDurableFile(dir, name string, data []byte) (string, error) {
	return writeMaterialFile(dir, name, data, true)
}

func writeMaterialFile(dir, name string, data []byte, keep bool) (string, error) {
	path := filepath.Join(dir, name)
	if keep {
		exists, err := installedFile(path)
		if err != nil {
			return "", err
		}
		if exists {
			return path, nil
		}
	}
	if err := writePrivate(path, data); err != nil {
		return "", err
	}
	return path, nil
}

func installedFile(path string) (bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Size() == 0 {
		return false, nil
	}
	return true, nil
}
