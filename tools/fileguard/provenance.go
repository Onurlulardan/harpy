package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type generatedManifest struct {
	Version int                  `json:"version"`
	Files   []generatedFileEntry `json:"files"`
}

type generatedFileEntry struct {
	Path         string `json:"path"`
	Generator    string `json:"generator"`
	InputSHA256  string `json:"input_sha256"`
	OutputSHA256 string `json:"output_sha256"`
}

func loadGeneratedManifest(root string) (map[string]generatedFileEntry, error) {
	path := filepath.Join(root, "contracts", "generated.yaml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]generatedFileEntry{}, nil
	}
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest generatedManifest
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("decode generated manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("decode generated manifest: trailing JSON value")
	}
	if manifest.Version != 1 {
		return nil, fmt.Errorf("generated manifest version must be 1")
	}
	if manifest.Files == nil {
		return nil, fmt.Errorf("generated manifest files must be an array")
	}
	entries := make(map[string]generatedFileEntry, len(manifest.Files))
	for _, entry := range manifest.Files {
		canonical, err := canonicalManifestPath(entry.Path)
		if err != nil {
			return nil, err
		}
		if entry.Generator == "" || !validSHA256(entry.InputSHA256) || !validSHA256(entry.OutputSHA256) {
			return nil, fmt.Errorf("generated entry %s has invalid provenance", canonical)
		}
		if _, exists := entries[canonical]; exists {
			return nil, fmt.Errorf("duplicate generated entry %s", canonical)
		}
		entry.Path = canonical
		entries[canonical] = entry
	}
	return entries, nil
}

func canonicalManifestPath(path string) (string, error) {
	clean := filepath.ToSlash(filepath.Clean(path))
	if clean == "." || filepath.IsAbs(path) || clean == ".." ||
		strings.HasPrefix(clean, "../") || clean != path {
		return "", fmt.Errorf("generated path %q is not canonical", path)
	}
	return clean, nil
}

func verifyGenerated(path string, entry generatedFileEntry) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(data)
	if hex.EncodeToString(digest[:]) != entry.OutputSHA256 {
		return fmt.Errorf("output hash does not match provenance manifest")
	}
	return nil
}

func validSHA256(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && value == strings.ToLower(value)
}
