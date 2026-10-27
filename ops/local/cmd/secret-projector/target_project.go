package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type ownershipChanger func(string, int, int) error

func projectTarget(
	root string,
	source verifiedSource,
	target targetSpec,
	chown ownershipChanger,
) error {
	if err := requireRealDirectory(root); err != nil {
		return fmt.Errorf("target %s: %w", target.name, err)
	}
	expected := markerFor(source, target)
	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("target %s: read volume: %w", target.name, err)
	}
	if len(entries) > 0 {
		complete := filepath.Join(root, completeMarker)
		if _, err := os.Lstat(complete); err == nil {
			if err := validateProjection(root, expected, target, chown); err != nil {
				if ok, _ := markerMatches(root, expected); !ok {
					if err := clearDirectory(root); err != nil {
						return fmt.Errorf("target %s: remove stale projection: %w", target.name, err)
					}
				} else {
					return err
				}
			} else {
				return nil
			}
		} else {
			pending := filepath.Join(root, pendingMarker)
			if _, err := os.Lstat(pending); err == nil {
				if err := recoverPending(root, expected); err != nil {
					return fmt.Errorf("target %s: %w", target.name, err)
				}
			} else {
				return fmt.Errorf("target %s: volume is not an owned projection", target.name)
			}
		}
	}
	if err := createProjection(root, source, expected, target, chown); err != nil {
		return fmt.Errorf("target %s: %w", target.name, err)
	}
	return nil
}

func createProjection(
	root string,
	source verifiedSource,
	marker projectionMarker,
	target targetSpec,
	chown ownershipChanger,
) error {
	content, err := marshalMarker(marker)
	if err != nil {
		return err
	}
	if err := writeOwnedFile(filepath.Join(root, pendingMarker), content, 0o400,
		target.uid, target.gid, chown); err != nil {
		return err
	}
	for _, file := range target.files {
		expected := source.files[file.name]
		content, err := readRegularFile(filepath.Join(source.root, file.name), expected.Size+1)
		if err != nil || int64(len(content)) != expected.Size || !contentMatches(content, expected) {
			return fmt.Errorf("source artifact %s changed during projection", file.name)
		}
		if err := writeOwnedFile(filepath.Join(root, file.name), content, file.mode,
			target.uid, target.gid, chown); err != nil {
			return err
		}
	}
	if err := os.Rename(
		filepath.Join(root, pendingMarker),
		filepath.Join(root, completeMarker),
	); err != nil {
		return err
	}
	if err := os.Chmod(root, target.dir); err != nil {
		return err
	}
	return chown(root, target.uid, target.gid)
}

func requireRealDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("mounted path is not a real directory")
	}
	return nil
}

func clearDirectory(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(root, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}
