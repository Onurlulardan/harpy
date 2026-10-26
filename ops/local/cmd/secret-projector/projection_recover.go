package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func recoverPending(root string, expected projectionMarker) error {
	path := filepath.Join(root, pendingMarker)
	content, err := readRegularFile(path, 1<<20)
	if err != nil {
		return fmt.Errorf("read pending marker: %w", err)
	}
	actual, err := decodeMarker(content)
	if err != nil || !markersEqual(actual, expected) {
		return fmt.Errorf("pending projection does not match current source")
	}
	allowed := map[string]bool{pendingMarker: true}
	for _, file := range expected.Files {
		allowed[file.Name] = true
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !allowed[entry.Name()] || entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("pending projection contains an unowned entry")
		}
	}
	for _, entry := range entries {
		if err := os.Remove(filepath.Join(root, entry.Name())); err != nil {
			return fmt.Errorf("remove partial projection: %w", err)
		}
	}
	return nil
}
