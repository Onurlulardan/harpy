package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const recoveryUpgradeSourceRevision = 4

func upgradeRecoveryArtifactSet(
	output string,
	now time.Time,
) (ensureResult, bool, error) {
	legacy, intact, err := inspectManifestBackedSet(output, func(value setManifest) error {
		return validateRecoveryUpgradeManifest(value, now)
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
	if err := addRecoveryArtifacts(set, now); err != nil {
		return ensureResult{}, false, err
	}
	if err := addAuthorityArtifacts(set, now); err != nil {
		return ensureResult{}, false, err
	}
	if err := refreshAuthorityProfileEnvironments(set); err != nil {
		return ensureResult{}, false, err
	}
	if err := addTenantMigrationExecutorEnvironment(set); err != nil {
		return ensureResult{}, false, err
	}
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		return ensureResult{}, false, err
	}
	staged, backup := upgradePaths(output)
	if err := requireAbsentUpgradePath(staged); err != nil {
		return ensureResult{}, false, err
	}
	if err := requireAbsentUpgradePath(backup); err != nil {
		return ensureResult{}, false, err
	}
	if err := publishArtifactSet(staged, set, manifest); err != nil {
		return ensureResult{}, false, err
	}
	if err := os.Rename(output, backup); err != nil {
		_ = os.RemoveAll(staged)
		return ensureResult{}, false, fmt.Errorf("stage local config upgrade: %w", err)
	}
	if err := os.Rename(staged, output); err != nil {
		rollbackErr := os.Rename(backup, output)
		return ensureResult{}, false, fmt.Errorf(
			"publish local config upgrade: %w; rollback: %v", err, rollbackErr)
	}
	if err := os.RemoveAll(backup); err != nil {
		return ensureResult{}, false, fmt.Errorf("remove local config upgrade backup: %w", err)
	}
	return ensureResult{Generated: true, Output: output, Manifest: manifest}, true, nil
}

func validateRecoveryUpgradeManifest(value setManifest, now time.Time) error {
	if value.ArtifactContractRevision != recoveryUpgradeSourceRevision ||
		(value.Profile != "" && value.Profile != string(profileLocal)) ||
		value.ProfileInputsSHA256 != "" || value.LicenseKind != "" {
		return fmt.Errorf("manifest is not eligible for recovery upgrade")
	}
	if err := validateManifestLifecycle(value, now, false); err != nil {
		return err
	}
	if err := validateManifestFileCatalog(value.Files); err != nil {
		return err
	}
	expected := expectedArtifactNames()
	for _, name := range recoveryArtifactNames {
		expected = removeArtifactName(expected, name)
	}
	for _, name := range authorityArtifactNames {
		expected = removeArtifactName(expected, name)
	}
	expected = removeArtifactName(expected, "tenant-migration-executor.env")
	sort.Strings(expected)
	if len(value.Files) != len(expected) {
		return fmt.Errorf("recovery upgrade source catalog differs")
	}
	for index, file := range value.Files {
		if file.Name != expected[index] {
			return fmt.Errorf("recovery upgrade source catalog differs")
		}
	}
	return nil
}

func recoverArtifactUpgrade(output string, now time.Time) error {
	staged, backup := upgradePaths(output)
	_, backupErr := os.Lstat(backup)
	if os.IsNotExist(backupErr) {
		if _, stagedErr := os.Lstat(staged); stagedErr == nil {
			if _, outputErr := os.Lstat(output); outputErr == nil {
				return os.RemoveAll(staged)
			}
			return fmt.Errorf("orphaned local config upgrade stage requires operator review")
		}
		return nil
	}
	if backupErr != nil {
		return backupErr
	}
	if _, current, _ := inspectArtifactSet(output, now, false); current {
		_ = os.RemoveAll(staged)
		return os.RemoveAll(backup)
	}
	if _, outputErr := os.Lstat(output); !os.IsNotExist(outputErr) {
		return fmt.Errorf("local config upgrade recovery found an invalid output")
	}
	if _, stagedCurrent, _ := inspectArtifactSet(staged, now, false); stagedCurrent {
		if err := os.Rename(staged, output); err != nil {
			return err
		}
		return os.RemoveAll(backup)
	}
	_ = os.RemoveAll(staged)
	return os.Rename(backup, output)
}

func upgradePaths(output string) (string, string) {
	return output + ".revision-7-staged", output + ".revision-prior-backup"
}

func requireAbsentUpgradePath(path string) error {
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("local config upgrade path %s already exists", filepath.Base(path))
}
