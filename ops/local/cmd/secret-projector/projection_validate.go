package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

func validateProjection(
	root string,
	expected projectionMarker,
	target targetSpec,
	chown ownershipChanger,
) error {
	content, err := readRegularFile(filepath.Join(root, completeMarker), 1<<20)
	if err != nil {
		return fmt.Errorf("read projection marker: %w", err)
	}
	actual, err := decodeMarker(content)
	if err != nil || !markersEqual(actual, expected) {
		return fmt.Errorf("projection marker does not match current source")
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != len(target.files)+1 {
		return fmt.Errorf("projection catalog is invalid")
	}
	allowed := map[string]bool{completeMarker: true}
	for _, file := range target.files {
		allowed[file.name] = true
		manifestFile := expectedFile(expected, file.name)
		content, err := readRegularFile(filepath.Join(root, file.name), manifestFile.Size+1)
		if err != nil || int64(len(content)) != manifestFile.Size ||
			!contentMatches(content, manifestFile) {
			return fmt.Errorf("projected artifact %s is invalid", file.name)
		}
		path := filepath.Join(root, file.name)
		if err := os.Chmod(path, file.mode); err != nil {
			return err
		}
		if err := chown(path, target.uid, target.gid); err != nil {
			return err
		}
	}
	for _, entry := range entries {
		if !allowed[entry.Name()] || entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("projection contains an unowned entry")
		}
	}
	markerPath := filepath.Join(root, completeMarker)
	if err := os.Chmod(markerPath, 0o400); err != nil {
		return err
	}
	if err := chown(markerPath, target.uid, target.gid); err != nil {
		return err
	}
	if err := os.Chmod(root, target.dir); err != nil {
		return err
	}
	return chown(root, target.uid, target.gid)
}

func expectedFile(marker projectionMarker, name string) sourceFile {
	for _, file := range marker.Files {
		if file.Name == name {
			return file
		}
	}
	return sourceFile{}
}

func contentMatches(content []byte, expected sourceFile) bool {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:]) == expected.SHA256
}

func markerMatches(root string, expected projectionMarker) (bool, error) {
	content, err := readRegularFile(filepath.Join(root, completeMarker), 1<<20)
	if err != nil {
		return false, err
	}
	actual, err := decodeMarker(content)
	if err != nil {
		return false, err
	}
	return markersEqual(actual, expected), nil
}
