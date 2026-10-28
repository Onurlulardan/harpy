package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

const foundationMigrationDigest = "f9fef3e093296fb402caed0a48415560aff29224cdbc62b56926f1209718b266"

func loadFoundationMigrationArtifact() (artifact, error) {
	root, err := findRepositoryRoot()
	if err != nil {
		return artifact{}, err
	}
	path := filepath.Join(
		root, releaseArtifactDirectory,
		foundationMigrationDigest+".migration-manifest.json",
	)
	document, err := os.ReadFile(path)
	if err != nil {
		return artifact{}, fmt.Errorf("read foundation migration artifact: %w", err)
	}
	verified, err := attestation.VerifyMigrationManifestArtifact(document)
	if err != nil || verified.DigestSHA256() != foundationMigrationDigest {
		return artifact{}, fmt.Errorf("verify foundation migration artifact: %w", err)
	}
	manifest := verified.Value()
	if len(manifest.Entries) != 138 {
		return artifact{}, fmt.Errorf("foundation migration entry count differs")
	}
	frontier := manifest.Entries[len(manifest.Entries)-1]
	if frontier.Version != 20260714014100 || frontier.TargetDatabaseGeneration != 2 ||
		frontier.NextMinimumRuntimeGeneration != 1 {
		return artifact{}, fmt.Errorf("foundation migration frontier differs")
	}
	scopes := map[string]bool{}
	required := map[int64]bool{
		20260714005900: false, 20260714010000: false,
		20260714012910: false, 20260714012920: false, 20260714012930: false,
	}
	for _, entry := range manifest.Entries {
		scopes[entry.Scope] = true
		if _, exists := required[entry.Version]; exists {
			required[entry.Version] = true
		}
	}
	if !scopes["platform"] || !scopes["tenant"] {
		return artifact{}, fmt.Errorf("foundation migration scope set is incomplete")
	}
	for _, present := range required {
		if !present {
			return artifact{}, fmt.Errorf("foundation migration finalization set is incomplete")
		}
	}
	return artifact{
		name: filepath.Base(path), digest: foundationMigrationDigest, document: document,
	}, nil
}
