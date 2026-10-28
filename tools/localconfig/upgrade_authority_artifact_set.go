package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const authorityUpgradeSourceRevision = 5

func upgradeAuthorityArtifactSet(
	output string,
	now time.Time,
) (ensureResult, bool, error) {
	legacy, intact, err := inspectManifestBackedSet(output, func(value setManifest) error {
		return validateAuthorityUpgradeManifest(value, now)
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
	if err = addAuthorityArtifacts(set, now); err != nil {
		return ensureResult{}, false, err
	}
	if err = refreshAuthorityProfileEnvironments(set); err != nil {
		return ensureResult{}, false, err
	}
	if err = addTenantMigrationExecutorEnvironment(set); err != nil {
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
		return ensureResult{}, false, fmt.Errorf("stage authority config upgrade: %w", err)
	}
	if err = os.Rename(staged, output); err != nil {
		rollbackErr := os.Rename(backup, output)
		return ensureResult{}, false, fmt.Errorf(
			"publish authority config upgrade: %w; rollback: %v", err, rollbackErr)
	}
	if err = os.RemoveAll(backup); err != nil {
		return ensureResult{}, false, err
	}
	return ensureResult{Generated: true, Output: output, Manifest: manifest}, true, nil
}

func validateAuthorityUpgradeManifest(value setManifest, now time.Time) error {
	if value.ArtifactContractRevision != authorityUpgradeSourceRevision ||
		(value.Profile != "" && value.Profile != string(profileLocal)) ||
		value.ProfileInputsSHA256 != "" || value.LicenseKind != "" {
		return fmt.Errorf("manifest is not eligible for authority upgrade")
	}
	if err := validateManifestLifecycle(value, now, false); err != nil {
		return err
	}
	if err := validateManifestFileCatalog(value.Files); err != nil {
		return err
	}
	expected := expectedArtifactNames()
	for _, name := range []string{
		"local-authority-ack-signing-seed",
		"local-authority-release-signing-seed",
		"tenant-migration-executor.env",
	} {
		expected = removeArtifactName(expected, name)
	}
	sort.Strings(expected)
	if len(value.Files) != len(expected) {
		return fmt.Errorf("authority upgrade source catalog differs")
	}
	for index, file := range value.Files {
		if file.Name != expected[index] {
			return fmt.Errorf("authority upgrade source catalog differs")
		}
	}
	return nil
}
