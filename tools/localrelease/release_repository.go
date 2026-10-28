package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/Onurlulardan/harpy/internal/migrationbundle"
)

const releaseArtifactDirectory = "core/internal/platform/migration/release"

func findRepositoryRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for {
		module, readErr := os.ReadFile(filepath.Join(current, "go.mod"))
		line, _, _ := strings.Cut(string(module), "\n")
		if readErr == nil && strings.TrimSpace(line) == "module github.com/Onurlulardan/harpy" {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repository root not found")
		}
		current = parent
	}
}

func loadMigrationArtifact() (artifact, error) {
	root, err := findRepositoryRoot()
	if err != nil {
		return artifact{}, err
	}
	document, err := migrationbundle.CanonicalFresh(root)
	if err != nil {
		return artifact{}, err
	}
	verified, err := attestation.VerifyMigrationManifestArtifact(document)
	if err != nil {
		return artifact{}, fmt.Errorf("verify canonical fresh migration: %w", err)
	}
	if verified.DigestSHA256() != migrationbundle.Digest(document) {
		return artifact{}, fmt.Errorf("canonical fresh migration digest differs")
	}
	return artifact{
		name: "migration-manifest.json", digest: verified.DigestSHA256(), document: document,
	}, nil
}
