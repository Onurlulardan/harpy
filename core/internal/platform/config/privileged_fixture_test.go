package config

import (
	"path/filepath"
	"testing"
)

func privilegedEnvironment(t *testing.T, current profile) []string {
	t.Helper()
	directory := t.TempDir()
	receiptKey := "HARPY_BOOTSTRAP_RECEIPT_PATH"
	dsns := []string{
		"HARPY_BOOTSTRAP_MAINTENANCE_DSN=postgresql://harpy_cluster_admin:secret@db.invalid/postgres?sslmode=disable",
		"HARPY_BOOTSTRAP_TARGET_DSN=postgresql://harpy_cluster_admin:secret@db.invalid/app?sslmode=disable",
	}
	artifacts := []string{
		"HARPY_ROLES_MANIFEST_PATH=" + filepath.Join(directory, "roles.json"),
		"HARPY_PRIVILEGES_MANIFEST_PATH=" + filepath.Join(directory, "privileges.json"),
		"HARPY_OWNERSHIP_MANIFEST_PATH=" + filepath.Join(directory, "ownership.json"),
		"HARPY_BOOTSTRAP_OPERATION_PATH=" + filepath.Join(directory, "bootstrap-operation.json"),
	}
	if current == profileExtensionAdmin {
		receiptKey = "HARPY_EXTENSION_RECEIPT_PATH"
		dsns = []string{
			"HARPY_EXTENSION_ADMIN_POSTGRES_DSN=postgresql://harpy_extension_admin:secret@db.invalid/app?sslmode=disable",
		}
		artifacts = []string{
			"HARPY_EXTENSION_MANIFEST_PATH=" + filepath.Join(directory, "extensions.json"),
			"HARPY_OWNERSHIP_MANIFEST_PATH=" + filepath.Join(directory, "ownership.json"),
			"HARPY_EXTENSION_OPERATION_PATH=" + filepath.Join(directory, "extension-operation.json"),
		}
	}
	entries := []string{
		"HARPY_ENVIRONMENT=development",
		"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_RELEASE_MANIFEST_PATH=" + filepath.Join(directory, "release.json"),
		"HARPY_RELEASE_SIGNATURE_PATH=" + filepath.Join(directory, "release.sig"),
		"HARPY_RELEASE_TRUST_ROOT_PATH=" + filepath.Join(directory, "trust.pem"),
		"HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH=" + filepath.Join(directory, "authority.json"),
		"HARPY_BOOTSTRAP_AUTHORITY_SIGNATURE_PATH=" + filepath.Join(directory, "authority.sig"),
		"HARPY_CREDENTIAL_LEASE_ID=55555555-5555-4555-8555-555555555555",
		receiptKey + "=" + filepath.Join(directory, "receipt.json"),
		"HARPY_ATTESTATION_SIGNING_KEY=ephemeral-test-key",
	}
	entries = append(entries, dsns...)
	return append(entries, artifacts...)
}

func migratorEnvironment(t *testing.T) []string {
	t.Helper()
	directory := t.TempDir()
	return append(validEnvironment(),
		"HARPY_RELEASE_MANIFEST_PATH="+filepath.Join(directory, "release.json"),
		"HARPY_RELEASE_SIGNATURE_PATH="+filepath.Join(directory, "release.sig"),
		"HARPY_RELEASE_TRUST_ROOT_PATH="+filepath.Join(directory, "trust.pem"),
		"HARPY_BOOTSTRAP_RECEIPT_PATH="+filepath.Join(directory, "bootstrap.json"),
		"HARPY_EXTENSION_RECEIPT_PATH="+filepath.Join(directory, "extension.json"),
		"HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH="+filepath.Join(directory, "authority.json"),
		"HARPY_BOOTSTRAP_AUTHORITY_SIGNATURE_PATH="+filepath.Join(directory, "authority.sig"),
		"HARPY_CREDENTIAL_CLEANUP_RECEIPT_PATH="+filepath.Join(directory, "cleanup.json"),
		"HARPY_CATALOG_PACKAGE_SET_PATH="+filepath.Join(directory, "catalog-package-set.json"),
		catalogPackageKeyringEnvironmentEntry(),
		"HARPY_MIGRATION_MANIFEST_PATH="+filepath.Join(directory, "migrations.json"),
		"HARPY_EXTENSION_MANIFEST_PATH="+filepath.Join(directory, "extensions.json"),
		"HARPY_ROLES_MANIFEST_PATH="+filepath.Join(directory, "roles.json"),
		"HARPY_PRIVILEGES_MANIFEST_PATH="+filepath.Join(directory, "privileges.json"),
		"HARPY_OWNERSHIP_MANIFEST_PATH="+filepath.Join(directory, "ownership.json"),
		"HARPY_BOOTSTRAP_OPERATION_PATH="+filepath.Join(directory, "bootstrap-operation.json"),
		"HARPY_EXTENSION_OPERATION_PATH="+filepath.Join(directory, "extension-operation.json"),
		"HARPY_CREDENTIAL_CLEANUP_OPERATION_PATH="+filepath.Join(directory, "cleanup-operation.json"),
	)
}
