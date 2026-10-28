package main

import (
	"fmt"
	"os"
	"time"
)

func resetArtifactSet(output string, now time.Time) (bool, string, error) {
	resolved, err := resolveOutputPath(output)
	if err != nil {
		return false, "", err
	}
	if _, err := os.Lstat(resolved); os.IsNotExist(err) {
		return false, resolved, nil
	} else if err != nil {
		return false, resolved, fmt.Errorf("inspect reset target: %w", err)
	}
	_, intact, err := inspectResettableArtifactSet(resolved, now)
	if err != nil {
		return false, resolved, err
	}
	if !intact {
		return false, resolved, fmt.Errorf(
			"reset refused: artifact catalog or hashes do not match the tool-owned manifest",
		)
	}
	if err := os.RemoveAll(resolved); err != nil {
		return false, resolved, fmt.Errorf("remove verified artifact set: %w", err)
	}
	return true, resolved, nil
}
