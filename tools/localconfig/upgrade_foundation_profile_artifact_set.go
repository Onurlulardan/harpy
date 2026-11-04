package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const foundationProfileUpgradeSourceRevision = 12

func upgradeFoundationProfileArtifactSet(
	output string, now time.Time,
) (ensureResult, bool, error) {
	legacy, intact, err := inspectManifestBackedSet(output, func(value setManifest) error {
		return validateFoundationProfileUpgradeManifest(value, now)
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
	if err = rewriteFoundationProfileDSNs(set); err != nil {
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
		return ensureResult{}, false, fmt.Errorf("stage foundation profile upgrade: %w", err)
	}
	if err = os.Rename(staged, output); err != nil {
		rollbackErr := os.Rename(backup, output)
		return ensureResult{}, false, fmt.Errorf(
			"publish foundation profile upgrade: %w; rollback: %v", err, rollbackErr)
	}
	if err = os.RemoveAll(backup); err != nil {
		return ensureResult{}, false, err
	}
	return ensureResult{Generated: true, Output: output, Manifest: manifest}, true, nil
}

func rewriteFoundationProfileDSNs(set *artifactSet) error {
	apiPassword := string(set.files["postgres-api-password.txt"])
	if apiPassword == "" {
		return fmt.Errorf("foundation profile upgrade is missing the api password")
	}
	dsn := []byte(postgresDSN("harpy_api", apiPassword, "harpy"))
	set.files["postgres-runtime-dsn.txt"] = append([]byte(nil), dsn...)
	set.files["postgres-backup-dsn.txt"] = append([]byte(nil), dsn...)
	return nil
}

func validateFoundationProfileUpgradeManifest(value setManifest, now time.Time) error {
	if value.ArtifactContractRevision != foundationProfileUpgradeSourceRevision ||
		(value.Profile != "" && value.Profile != string(profileLocal)) ||
		value.ProfileInputsSHA256 != "" || value.LicenseKind != "" {
		return fmt.Errorf("manifest is not eligible for foundation profile upgrade")
	}
	if err := validateManifestLifecycle(value, now, false); err != nil {
		return err
	}
	return validateCurrentManifestCatalog(value.Files, profileLocal)
}
