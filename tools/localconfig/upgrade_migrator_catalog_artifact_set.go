package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const migratorCatalogUpgradeSourceRevision = 10

func migratorCatalogPackageEnvironment() profileEnvironment {
	return profileEnvironment{
		"HARPY_CATALOG_PACKAGE_SET_PATH": releaseRoot + "catalog-package-set.json",
		"HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING_FILE": secretRoot +
			"catalog-package-public-keyring.json",
	}
}

func refreshMigratorCatalogEnvironment(set *artifactSet) error {
	values, err := mergeProfileEnvironment(
		set.files["migrator.env"], migratorCatalogPackageEnvironment(),
	)
	if err != nil {
		return fmt.Errorf("refresh migrator.env: %w", err)
	}
	set.files["migrator.env"] = values
	return nil
}

func upgradeMigratorCatalogArtifactSet(
	output string,
	now time.Time,
	options generationOptions,
) (ensureResult, bool, error) {
	legacy, intact, err := inspectManifestBackedSet(output, func(value setManifest) error {
		return validateMigratorCatalogUpgradeManifest(value, now, options)
	})
	if err != nil || !intact {
		return ensureResult{}, false, err
	}
	set := newArtifactSetFor(options)
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
	if err = refreshMigratorCatalogEnvironment(set); err != nil {
		return ensureResult{}, false, err
	}
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		return ensureResult{}, false, err
	}
	if err = publishMigratorCatalogUpgrade(output, set, manifest); err != nil {
		return ensureResult{}, false, err
	}
	return ensureResult{Generated: true, Output: output, Manifest: manifest}, true, nil
}

func validateMigratorCatalogUpgradeManifest(
	value setManifest,
	now time.Time,
	options generationOptions,
) error {
	if value.ArtifactContractRevision != migratorCatalogUpgradeSourceRevision {
		return fmt.Errorf("manifest is not eligible for migrator catalog upgrade")
	}
	if err := validateManifestLifecycle(value, now, false); err != nil {
		return err
	}
	if options.profile != profileLocal ||
		(value.Profile != "" && value.Profile != string(profileLocal)) ||
		value.ProfileInputsSHA256 != "" || value.LicenseKind != "" {
		return fmt.Errorf("migrator catalog upgrade source profile differs")
	}
	return validateCurrentManifestCatalog(value.Files, options.profile)
}

func publishMigratorCatalogUpgrade(
	output string,
	set *artifactSet,
	manifest setManifest,
) error {
	staged, backup := upgradePaths(output)
	if err := requireAbsentUpgradePath(staged); err != nil {
		return err
	}
	if err := requireAbsentUpgradePath(backup); err != nil {
		return err
	}
	if err := publishArtifactSet(staged, set, manifest); err != nil {
		return err
	}
	if err := os.Rename(output, backup); err != nil {
		_ = os.RemoveAll(staged)
		return fmt.Errorf("stage migrator catalog upgrade: %w", err)
	}
	if err := os.Rename(staged, output); err != nil {
		rollbackErr := os.Rename(backup, output)
		return fmt.Errorf(
			"publish migrator catalog upgrade: %w; rollback: %v", err, rollbackErr,
		)
	}
	if err := os.RemoveAll(backup); err != nil {
		return err
	}
	return nil
}
