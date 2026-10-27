package main

import (
	"fmt"
	"os"
)

func initializeStateVolume(
	path string,
	target targetSpec,
	chown ownershipChanger,
) error {
	if err := requireRealDirectory(path); err != nil {
		return fmt.Errorf("state target %s: %w", target.name, err)
	}
	if err := os.Chmod(path, target.dir); err != nil {
		return fmt.Errorf("state target %s: set mode: %w", target.name, err)
	}
	if err := chown(path, target.uid, target.gid); err != nil {
		return fmt.Errorf("state target %s: set owner: %w", target.name, err)
	}
	return nil
}
