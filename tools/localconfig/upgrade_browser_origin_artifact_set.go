package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const browserOriginUpgradeSourceRevision = 8

var (
	legacyLocalBrowserOrigins = []byte(
		"HARPY_IAM_ALLOWED_ORIGINS=[\"http://localhost:5173\"]\n",
	)
	currentLocalBrowserOrigins = []byte(
		"HARPY_IAM_ALLOWED_ORIGINS=" +
			"[\"http://localhost:5173\",\"http://127.0.0.1:5173\"]\n",
	)
)

func upgradeBrowserOriginArtifactSet(
	output string,
	now time.Time,
	options generationOptions,
) (ensureResult, bool, error) {
	legacy, intact, err := inspectManifestBackedSet(output, func(value setManifest) error {
		return validateBrowserOriginUpgradeManifest(value, now, options)
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
	if options.profile == profileLocal {
		apiEnvironment := set.files["api.env"]
		if bytes.Count(apiEnvironment, legacyLocalBrowserOrigins) != 1 ||
			bytes.Contains(apiEnvironment, currentLocalBrowserOrigins) {
			return ensureResult{}, false, fmt.Errorf(
				"local API environment is not eligible for browser origin upgrade",
			)
		}
		set.files["api.env"] = bytes.Replace(
			apiEnvironment, legacyLocalBrowserOrigins, currentLocalBrowserOrigins, 1,
		)
	}
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		return ensureResult{}, false, err
	}
	if err = publishBrowserOriginUpgrade(output, set, manifest); err != nil {
		return ensureResult{}, false, err
	}
	return ensureResult{Generated: true, Output: output, Manifest: manifest}, true, nil
}

func validateBrowserOriginUpgradeManifest(
	value setManifest,
	now time.Time,
	options generationOptions,
) error {
	if value.ArtifactContractRevision != browserOriginUpgradeSourceRevision {
		return fmt.Errorf("manifest is not eligible for browser origin upgrade")
	}
	if err := validateManifestLifecycle(value, now, false); err != nil {
		return err
	}
	if options.profile == profileLocal {
		if (value.Profile != "" && value.Profile != string(profileLocal)) ||
			value.ProfileInputsSHA256 != "" || value.LicenseKind != "" {
			return fmt.Errorf("browser origin upgrade source profile differs")
		}
	} else if value.Profile != string(profileForgeGuard) ||
		value.ProfileInputsSHA256 != options.digest() ||
		value.LicenseKind != options.licenseKind {
		return fmt.Errorf("browser origin upgrade source profile differs")
	}
	return validateCurrentManifestCatalog(
		value.Files, options.profile, options.licenseKind,
	)
}

func publishBrowserOriginUpgrade(
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
		return fmt.Errorf("stage browser origin upgrade: %w", err)
	}
	if err := os.Rename(staged, output); err != nil {
		rollbackErr := os.Rename(backup, output)
		return fmt.Errorf(
			"publish browser origin upgrade: %w; rollback: %v", err, rollbackErr,
		)
	}
	if err := os.RemoveAll(backup); err != nil {
		return err
	}
	return nil
}
