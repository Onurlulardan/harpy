package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func inspectArtifactSet(
	output string,
	now time.Time,
	ignoreExpiry bool,
) (setManifest, bool, error) {
	return inspectManifestBackedSet(output, func(manifest setManifest) error {
		return validateManifestDocument(manifest, now, ignoreExpiry)
	})
}

func inspectResettableArtifactSet(
	output string,
	now time.Time,
) (setManifest, bool, error) {
	return inspectManifestBackedSet(output, func(manifest setManifest) error {
		return validateResetManifestDocument(manifest, now)
	})
}

func inspectManifestBackedSet(
	output string,
	validate func(setManifest) error,
) (setManifest, bool, error) {
	info, err := os.Lstat(output)
	if os.IsNotExist(err) {
		return setManifest{}, false, nil
	}
	if err != nil {
		return setManifest{}, false, fmt.Errorf("inspect artifact set: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return setManifest{}, false, fmt.Errorf("artifact set is not a real directory")
	}
	manifestContent, err := os.ReadFile(filepath.Join(output, manifestFileName))
	if os.IsNotExist(err) {
		return setManifest{}, false, nil
	}
	if err != nil {
		return setManifest{}, false, fmt.Errorf("read manifest: %w", err)
	}
	manifest, err := decodeSetManifest(manifestContent)
	if err != nil || validate(manifest) != nil {
		return setManifest{}, false, nil
	}
	entries, err := os.ReadDir(output)
	if err != nil {
		return setManifest{}, false, fmt.Errorf("read artifact set: %w", err)
	}
	if len(entries) != len(manifest.Files)+1 {
		return setManifest{}, false, nil
	}
	expected := make(map[string]bool, len(manifest.Files)+1)
	expected[manifestFileName] = true
	for _, file := range manifest.Files {
		expected[file.Name] = true
		if !artifactFileMatches(output, file) {
			return setManifest{}, false, nil
		}
	}
	for _, entry := range entries {
		if !expected[entry.Name()] || entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			return setManifest{}, false, nil
		}
	}
	return manifest, true, nil
}

func artifactFileMatches(output string, file manifestFile) bool {
	path := filepath.Join(output, file.Name)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 ||
		info.Size() != int64(file.Size) {
		return false
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:]) == file.SHA256
}
