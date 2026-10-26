package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type verifiedSource struct {
	root       string
	generation string
	files      map[string]sourceFile
}

func verifySource(root string, now time.Time) (verifiedSource, error) {
	return verifySourceFor(root, now, secretTargets())
}

func verifySourceFor(
	root string,
	now time.Time,
	targets []targetSpec,
) (verifiedSource, error) {
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return verifiedSource{}, fmt.Errorf("source is not a real directory")
	}
	manifestPath := filepath.Join(root, "manifest.json")
	content, err := readRegularFile(manifestPath, 1<<20)
	if err != nil {
		return verifiedSource{}, fmt.Errorf("read source manifest: %w", err)
	}
	manifest, err := decodeManifest(content)
	if err != nil {
		return verifiedSource{}, err
	}
	if err := validateManifest(manifest, now); err != nil {
		return verifiedSource{}, err
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != len(manifest.Files)+1 {
		return verifiedSource{}, fmt.Errorf("source catalog does not match manifest")
	}
	files := make(map[string]sourceFile, len(manifest.Files))
	for _, file := range manifest.Files {
		if err := verifySourceFile(root, file); err != nil {
			return verifiedSource{}, err
		}
		files[file.Name] = file
	}
	for _, entry := range entries {
		if entry.Name() == "manifest.json" {
			continue
		}
		if _, exists := files[entry.Name()]; !exists || entry.IsDir() ||
			entry.Type()&os.ModeSymlink != 0 {
			return verifiedSource{}, fmt.Errorf("source catalog contains an unowned entry")
		}
	}
	for _, target := range targets {
		for _, file := range target.files {
			if _, exists := files[file.name]; !exists {
				return verifiedSource{}, fmt.Errorf("target %s requires missing artifact %s",
					target.name, file.name)
			}
		}
	}
	return verifiedSource{root: root, generation: manifest.GenerationID, files: files}, nil
}

func verifySourceFile(root string, expected sourceFile) error {
	content, err := readRegularFile(filepath.Join(root, expected.Name), expected.Size+1)
	if err != nil || int64(len(content)) != expected.Size {
		return fmt.Errorf("source artifact %s is invalid", expected.Name)
	}
	digest := sha256.Sum256(content)
	if hex.EncodeToString(digest[:]) != expected.SHA256 {
		return fmt.Errorf("source artifact %s hash does not match", expected.Name)
	}
	return nil
}

func readRegularFile(path string, limit int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("not a regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, limit))
	if err != nil || int64(len(content)) >= limit {
		return nil, fmt.Errorf("file exceeds its allowed size")
	}
	return content, nil
}
