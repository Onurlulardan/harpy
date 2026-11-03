package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var errFrontendCheckoutAbsent = errors.New("sibling frontend-v3 checkout is absent")

func frontendCheckoutPresent(root string) (bool, error) {
	present := true
	for _, name := range []string{"package.json", "yarn.lock"} {
		_, err := os.Stat(filepath.Join(root, name))
		if err == nil {
			continue
		}
		if os.IsNotExist(err) {
			present = false
			continue
		}
		return false, fmt.Errorf("frontend-v3 %s is unavailable: %w", name, err)
	}
	return present, nil
}
