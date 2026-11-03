package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const foundationCredentialUpgradeSourceRevision = 11

var foundationCredentialFiles = []string{
	"initialize.env",
	"postgres-api-dsn.txt",
	"postgres-api-password.txt",
	"postgres-goose-dsn.txt",
	"postgres-goose-password.txt",
	"postgres-worker-dsn.txt",
	"postgres-worker-password.txt",
}

func upgradeFoundationCredentialArtifactSet(
	output string, now time.Time,
) (ensureResult, bool, error) {
	legacy, intact, err := inspectManifestBackedSet(output, func(value setManifest) error {
		return validateFoundationCredentialUpgradeManifest(value, now)
	})
	if err != nil || !intact {
		return ensureResult{}, false, err
	}
	set := newArtifactSet()
	for _, file := range legacy.Files {
		content, readErr := os.ReadFile(filepath.Join(output, file.Name))
		if readErr != nil {
			return ensureResult{}, false, readErr
		}
		set.files[file.Name] = content
	}
	set.expiresAt, err = parseCanonicalTimestamp(legacy.ExpiresAt)
	if err != nil {
		return ensureResult{}, false, err
	}
	if err = addFoundationCredentialArtifacts(set); err != nil {
		return ensureResult{}, false, err
	}
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		return ensureResult{}, false, err
	}
	staged, backup := upgradePaths(output)
	if err = requireAbsentUpgradePath(staged); err != nil {
		return ensureResult{}, false, err
	}
	if err = requireAbsentUpgradePath(backup); err != nil {
		return ensureResult{}, false, err
	}
	if err = publishArtifactSet(staged, set, manifest); err != nil {
		return ensureResult{}, false, err
	}
	if err = os.Rename(output, backup); err != nil {
		_ = os.RemoveAll(staged)
		return ensureResult{}, false, fmt.Errorf("stage foundation credential upgrade: %w", err)
	}
	if err = os.Rename(staged, output); err != nil {
		rollbackErr := os.Rename(backup, output)
		return ensureResult{}, false, fmt.Errorf(
			"publish foundation credential upgrade: %w; rollback: %v", err, rollbackErr)
	}
	if err = os.RemoveAll(backup); err != nil {
		return ensureResult{}, false, err
	}
	return ensureResult{Generated: true, Output: output, Manifest: manifest}, true, nil
}

func addFoundationCredentialArtifacts(set *artifactSet) error {
	passwords, err := newConnectionPasswords()
	if err != nil {
		return err
	}
	files := map[string]string{
		"postgres-goose-password.txt":  passwords.goose,
		"postgres-api-password.txt":    passwords.api,
		"postgres-worker-password.txt": passwords.worker,
		"postgres-goose-dsn.txt": postgresDSN(
			"harpy_goose", passwords.goose, "harpy"),
		"postgres-api-dsn.txt": postgresDSN(
			"harpy_api", passwords.api, "harpy"),
		"postgres-worker-dsn.txt": postgresDSN(
			"harpy_worker", passwords.worker, "harpy"),
	}
	for name, value := range files {
		if _, exists := set.files[name]; exists {
			continue
		}
		if err := set.add(name, []byte(value)); err != nil {
			return err
		}
	}
	if _, exists := set.files["initialize.env"]; exists {
		return nil
	}
	content, err := renderProfileEnvironment(initializeEnvironment())
	if err != nil {
		return err
	}
	return set.add("initialize.env", content)
}

func validateFoundationCredentialUpgradeManifest(value setManifest, now time.Time) error {
	if value.ArtifactContractRevision != foundationCredentialUpgradeSourceRevision ||
		(value.Profile != "" && value.Profile != string(profileLocal)) ||
		value.ProfileInputsSHA256 != "" || value.LicenseKind != "" {
		return fmt.Errorf("manifest is not eligible for foundation credential upgrade")
	}
	if err := validateManifestLifecycle(value, now, false); err != nil {
		return err
	}
	if err := validateManifestFileCatalog(value.Files); err != nil {
		return err
	}
	expected := append([]string(nil), expectedArtifactNames()...)
	for _, name := range foundationCredentialFiles {
		expected = removeArtifactName(expected, name)
	}
	sort.Strings(expected)
	if len(value.Files) != len(expected) {
		return fmt.Errorf("foundation credential upgrade source catalog differs")
	}
	for index, file := range value.Files {
		if file.Name != expected[index] {
			return fmt.Errorf("foundation credential upgrade source catalog differs")
		}
	}
	return nil
}
