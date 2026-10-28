package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const (
	ownedMarkerName = ".harpy-localrelease-owned"
	ownedMarkerBody = "harpy-localrelease.v1\n"
)

type ownedOutput struct {
	target, staging, parent string
	replace                 bool
	committed               bool
}

func beginOwnedOutput(target string, replace bool) (*ownedOutput, error) {
	target = filepath.Clean(target)
	parent := filepath.Dir(target)
	if target == parent || filepath.Base(target) == "." {
		return nil, fmt.Errorf("output directory cannot be a filesystem root")
	}
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return nil, fmt.Errorf("create output parent: %w", err)
	}
	if _, err := os.Lstat(target); err == nil {
		if !replace {
			return nil, fmt.Errorf("output directory already exists")
		}
		if err := verifyOwnedDirectory(target, parent); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("inspect output directory: %w", err)
	} else if replace {
		return nil, fmt.Errorf("replacement output directory does not exist")
	}
	staging := filepath.Join(parent, "."+filepath.Base(target)+".staging-"+uuid.NewString())
	if err := os.Mkdir(staging, 0o700); err != nil {
		return nil, fmt.Errorf("create output staging directory: %w", err)
	}
	if err := os.WriteFile(
		filepath.Join(staging, ownedMarkerName), []byte(ownedMarkerBody), 0o600,
	); err != nil {
		_ = os.Remove(staging)
		return nil, fmt.Errorf("write output ownership marker: %w", err)
	}
	return &ownedOutput{
		target: target, staging: staging, parent: parent, replace: replace,
	}, nil
}

func (output *ownedOutput) abort() {
	if output == nil || output.committed {
		return
	}
	_ = removeOwnedDirectory(output.staging, output.parent)
}

func (output *ownedOutput) commit() error {
	if err := verifyOwnedDirectory(output.staging, output.parent); err != nil {
		return err
	}
	if output.replace {
		return output.commitReplacement()
	}
	if _, err := os.Lstat(output.target); err == nil {
		return fmt.Errorf("output directory appeared during publication")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect prior output: %w", err)
	}
	if err := os.Rename(output.staging, output.target); err != nil {
		return fmt.Errorf("publish owned output atomically: %w", err)
	}
	output.committed = true
	return nil
}

func (output *ownedOutput) commitReplacement() error {
	if err := verifyOwnedDirectory(output.target, output.parent); err != nil {
		return err
	}
	backup := filepath.Join(
		output.parent, "."+filepath.Base(output.target)+".previous-"+uuid.NewString(),
	)
	if err := os.Rename(output.target, backup); err != nil {
		return fmt.Errorf("stage prior owned output: %w", err)
	}
	if err := os.Rename(output.staging, output.target); err != nil {
		rollbackErr := os.Rename(backup, output.target)
		return fmt.Errorf("publish replacement output: %w; rollback: %v", err, rollbackErr)
	}
	output.committed = true
	if err := removeOwnedDirectory(backup, output.parent); err != nil {
		return fmt.Errorf("remove prior owned output: %w", err)
	}
	return nil
}

func verifyOwnedDirectory(path, expectedParent string) error {
	clean := filepath.Clean(path)
	if filepath.Dir(clean) != filepath.Clean(expectedParent) {
		return fmt.Errorf("owned output escapes expected parent")
	}
	info, err := os.Lstat(clean)
	if err != nil {
		return fmt.Errorf("inspect owned output: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("owned output is not a direct directory")
	}
	markerPath := filepath.Join(clean, ownedMarkerName)
	markerInfo, err := os.Lstat(markerPath)
	if err != nil || !markerInfo.Mode().IsRegular() || markerInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("output directory ownership marker is invalid")
	}
	marker, err := os.ReadFile(markerPath)
	if err != nil || !bytes.Equal(marker, []byte(ownedMarkerBody)) {
		return fmt.Errorf("output directory is not owned by localrelease")
	}
	return nil
}

func removeOwnedDirectory(path, expectedParent string) error {
	if err := verifyOwnedDirectory(path, expectedParent); err != nil {
		return err
	}
	return os.RemoveAll(filepath.Clean(path))
}
