package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/Onurlulardan/harpy/tools/internal/migrationmanifest"
)

const releaseArtifactDirectory = "core/internal/platform/migration/release"

func findRepositoryRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for {
		module, readErr := os.ReadFile(filepath.Join(current, "go.mod"))
		if readErr == nil && strings.HasPrefix(string(module), "module github.com/Onurlulardan/harpy\n") {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("backend-v2 repository root not found")
		}
		current = parent
	}
}

func loadMigrationArtifact() (artifact, error) {
	root, err := findRepositoryRoot()
	if err != nil {
		return artifact{}, err
	}
	inventory, err := migrationmanifest.LoadStorageInventory(root)
	if err != nil {
		return artifact{}, err
	}
	groups, err := migrationmanifest.CollectAll(root, inventory)
	if err != nil {
		return artifact{}, err
	}
	manifest, err := migrationmanifest.Build(inventory, groups)
	if err != nil {
		return artifact{}, err
	}
	directory := filepath.Join(root, releaseArtifactDirectory)
	if err := migrationmanifest.ValidateReleaseFrontierAdvance(directory, manifest); err != nil {
		return artifact{}, err
	}
	document, digest, err := migrationmanifest.Encode(manifest)
	if err != nil {
		return artifact{}, err
	}
	path := filepath.Join(directory, digest+".migration-manifest.json")
	embedded, err := os.ReadFile(path)
	if err != nil {
		return artifact{}, fmt.Errorf("read canonical migration artifact: %w", err)
	}
	if !bytes.Equal(document, embedded) {
		return artifact{}, fmt.Errorf("canonical migration artifact differs from repository sources")
	}
	verified, err := attestation.VerifyMigrationManifestArtifact(document)
	if err != nil || verified.DigestSHA256() != digest {
		return artifact{}, fmt.Errorf("verify canonical migration artifact: %w", err)
	}
	return artifact{name: filepath.Base(path), digest: digest, document: document}, nil
}
