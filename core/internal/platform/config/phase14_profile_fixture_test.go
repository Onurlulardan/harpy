package config

import (
	"os"
	"path/filepath"
	"time"
)

func validOperatorInitializeEnvironment(mode DeploymentMode) []string {
	directory := phase14FixtureDirectory("operator-initialize")
	entries := append(validEnvironment(),
		"HARPY_INSTALLATION_MANIFEST_PATH="+phase14FixtureFile(directory, "installation.json", false),
		"HARPY_INITIAL_ADMIN_PASSWORD_FILE="+phase14FixtureFile(directory, "password", true),
		validIAMDirectoryEnvironmentEntry(time.Now().UTC().Truncate(time.Second)),
	)
	if mode == DeploymentSaaS {
		return append(entries,
			deploymentModeEnvironmentKey+"=saas", saasKMSKeyIDKey+"="+validAuthorityKMSKeyID,
			saasRegistryURLKey+"=https://authority.invalid",
			ackTrustBundleKey+"="+phase14FixtureFile(directory, "ack-trust.json", false),
			"HARPY_RELEASE_TRUST_ROOT_PATH="+phase14FixtureFile(directory, "release-trust.json", false),
		)
	}
	recoveryDirectory := filepath.Join(directory, "signing")
	if err := os.Mkdir(recoveryDirectory, 0o700); err != nil {
		panic(err)
	}
	return append(entries,
		"HARPY_RECOVERY_SIGNING_DIRECTORY="+recoveryDirectory,
		licenseTrustBundleKey+"="+phase14FixtureFile(directory, "license-trust.json", false),
		licenseSnapshotKey+"="+phase14FixtureFile(directory, "license.json", false),
	)
}

func validOperatorIdentityEnvironment() []string {
	directory := phase14FixtureDirectory("operator-identity")
	return append(validEnvironment(),
		"HARPY_RESTORE_RECEIPT_PATH="+phase14FixtureFile(directory, "receipt.json", false),
		"HARPY_RESTORE_SECURITY_EVIDENCE_PATH="+phase14FixtureFile(directory, "security.json", false),
		"HARPY_IAM_KEY_CUSTODY_MANIFEST_PATH="+phase14FixtureFile(directory, "custody.json", false),
		recoveryTrustBundleKey+"="+phase14FixtureFile(directory, "recovery-trust.json", false),
		ackTrustBundleKey+"="+phase14FixtureFile(directory, "ack-trust.json", false),
	)
}

func validOperatorLicenseEnvironment() []string {
	directory := phase14FixtureDirectory("operator-license")
	return append(validEnvironment(),
		"HARPY_RESTORE_RECEIPT_PATH="+phase14FixtureFile(directory, "receipt.json", false),
		recoveryTrustBundleKey+"="+phase14FixtureFile(directory, "recovery-trust.json", false),
		licenseTrustBundleKey+"="+phase14FixtureFile(directory, "license-trust.json", false),
		licenseSnapshotKey+"="+phase14FixtureFile(directory, "license.json", false),
	)
}

func validTenantMigrationEnvironment() []string {
	directory := phase14FixtureDirectory("tenant-migration")
	return append(validEnvironment(), releaseFixtureEntries(directory)...)
}

func validDatabaseBackupEnvironment() []string {
	directory := phase14FixtureDirectory("database-backup")
	entries := dedicatedDatabaseEnvironment(directory, "backup-dsn", "harpy_backup")
	entries = append(entries, releaseFixtureEntries(directory)...)
	return append(entries,
		"HARPY_RECOVERY_STORAGE_MANIFEST_PATH="+phase14FixtureFile(directory, "storage.json", false),
		recoveryTrustBundleKey+"="+phase14FixtureFile(directory, "recovery-trust.json", false),
		"HARPY_RECOVERY_MANIFEST_SIGNING_KEY_FILE="+phase14FixtureFile(directory, "manifest.key", true),
	)
}

func validDatabaseRestoreEnvironment() []string {
	directory := phase14FixtureDirectory("database-restore")
	entries := restoreDatabaseEnvironment(directory)
	entries = append(entries, releaseFixtureEntries(directory)...)
	entries = append(entries, restoreCommonFixtureEntries(directory)...)
	return append(entries,
		"HARPY_RESTORE_PLAN_PATH="+phase14FixtureFile(directory, "plan.json", false),
		"HARPY_DR_KEK_KEYRING_FILE="+phase14FixtureFile(directory, "dr-kek", true),
		"HARPY_RESTORE_RECEIPT_SIGNING_KEY_FILE="+phase14FixtureFile(directory, "receipt.key", true),
	)
}

func validDatabaseRehydrateEnvironment() []string {
	entries := validDatabaseRestoreEnvironment()
	entries = withoutEnvironmentKey(entries, "HARPY_RESTORE_PLAN_PATH")
	entries = withoutEnvironmentKey(entries, "HARPY_DR_KEK_KEYRING_FILE")
	entries = withoutEnvironmentKey(entries, "HARPY_RESTORE_RECEIPT_SIGNING_KEY_FILE")
	directory := filepath.Dir(environmentEntryValue(
		entries, "HARPY_RESTORE_CONTROL_POSTGRES_DSN_FILE",
	))
	return append(entries,
		"HARPY_REHYDRATE_REQUEST_PATH="+phase14FixtureFile(directory, "rehydrate.json", false),
		"HARPY_RECOVERY_KIT_FILE="+phase14FixtureFile(directory, "recovery.kit", true),
	)
}

func dedicatedDatabaseEnvironment(directory, name, principal string) []string {
	dsn := "postgresql://" + principal + "@example.invalid/harpy?sslmode=disable\n"
	return []string{
		"HARPY_ENVIRONMENT=development", "HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_" + map[string]string{"backup-dsn": "BACKUP"}[name] +
			"_POSTGRES_DSN_FILE=" + phase14FixtureFileContent(directory, name, dsn, true),
	}
}

func restoreDatabaseEnvironment(directory string) []string {
	control := "postgresql://harpy_restore@example.invalid/harpy?sslmode=disable\n"
	scratch := "postgresql://harpy_restore@example.invalid/harpy_restore_scratch?sslmode=disable\n"
	return []string{
		"HARPY_ENVIRONMENT=development",
		"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_RESTORE_CONTROL_POSTGRES_DSN_FILE=" +
			phase14FixtureFileContent(directory, "control-dsn", control, true),
		"HARPY_RESTORE_SCRATCH_POSTGRES_DSN_FILE=" +
			phase14FixtureFileContent(directory, "scratch-dsn", scratch, true),
	}
}

func releaseFixtureEntries(directory string) []string {
	return []string{
		"HARPY_RELEASE_MANIFEST_PATH=" + phase14FixtureFile(directory, "release.json", false),
		"HARPY_RELEASE_SIGNATURE_PATH=" + phase14FixtureFile(directory, "release.sig", false),
		"HARPY_RELEASE_TRUST_ROOT_PATH=" + phase14FixtureFile(directory, "release-trust.pem", false),
	}
}

func restoreCommonFixtureEntries(directory string) []string {
	return []string{
		"HARPY_RECOVERY_STORAGE_MANIFEST_PATH=" + phase14FixtureFile(directory, "storage.json", false),
		recoveryTrustBundleKey + "=" + phase14FixtureFile(directory, "recovery-trust.json", false),
		"HARPY_RESTORE_ORCHESTRATOR_TRUST_BUNDLE_PATH=" +
			phase14FixtureFile(directory, "orchestrator-trust.json", false),
		"HARPY_RESTORE_PROGRESS_PATH=" + filepath.Join(directory, "progress.json"),
		"HARPY_RESTORE_DRAIN_ACK_PATH=" + filepath.Join(directory, "drain-ack.json"),
		"HARPY_RESTORE_SWITCH_ACK_PATH=" + filepath.Join(directory, "switch-ack.json"),
		"HARPY_RESTORE_RECEIPT_PATH=" + filepath.Join(directory, "receipt.json"),
	}
}

func environmentEntryValue(entries []string, key string) string {
	for _, entry := range entries {
		if len(entry) > len(key) && entry[:len(key)+1] == key+"=" {
			return entry[len(key)+1:]
		}
	}
	return ""
}
