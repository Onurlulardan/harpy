package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
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
	document, err := runtimev1.ReadEmbeddedFreshMigrationManifest()
	if err != nil {
		return artifact{}, err
	}
	verified, err := attestation.VerifyMigrationManifestArtifact(document)
	if err != nil {
		return artifact{}, fmt.Errorf("verify canonical fresh migration: %w", err)
	}
	sum := sha256.Sum256(document)
	if verified.DigestSHA256() != hex.EncodeToString(sum[:]) {
		return artifact{}, fmt.Errorf("canonical fresh migration digest differs")
	}
	return artifact{
		name: "migration-manifest.json", digest: verified.DigestSHA256(), document: document,
	}, nil
}
