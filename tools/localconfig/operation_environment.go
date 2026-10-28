package main

import "fmt"

const releaseRoot = "/run/release/current/"

func schemaExecutorEnvironment(set *artifactSet) profileEnvironment {
	values := baseRuntimeEnvironmentFor(secretRoot + "postgres-migrator-dsn.txt")
	values["HARPY_MANAGEMENT_LISTEN_ADDR"] = "0.0.0.0:9000"
	return values
}

func migratorEnvironment(set *artifactSet) profileEnvironment {
	values := baseRuntimeEnvironmentFor(secretRoot + "postgres-migrator-dsn.txt")
	values["HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING_FILE"] =
		secretRoot + "catalog-package-public-keyring.json"
	paths := map[string]string{
		"HARPY_RELEASE_MANIFEST_PATH":                      "release.json",
		"HARPY_RELEASE_SIGNATURE_PATH":                     "release.sig",
		"HARPY_RELEASE_TRUST_ROOT_PATH":                    "trust-root.json",
		"HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH":          "bootstrap-authority.json",
		"HARPY_BOOTSTRAP_AUTHORITY_SIGNATURE_PATH":         "bootstrap-authority.sig",
		"HARPY_BOOTSTRAP_RECEIPT_PATH":                     "bootstrap-receipt.json",
		"HARPY_EXTENSION_RECEIPT_PATH":                     "extension-receipt.json",
		"HARPY_CREDENTIAL_CLEANUP_RECEIPT_PATH":            "cleanup-receipt.json",
		"HARPY_CATALOG_PACKAGE_SET_PATH":                   "catalog-package-set.json",
		"HARPY_MIGRATION_MANIFEST_PATH":                    "migration-manifest.json",
		"HARPY_EXTENSION_MANIFEST_PATH":                    "extension-manifest.json",
		"HARPY_ROLES_MANIFEST_PATH":                        "roles-manifest.json",
		"HARPY_PRIVILEGES_MANIFEST_PATH":                   "privileges-manifest.json",
		"HARPY_OWNERSHIP_MANIFEST_PATH":                    "ownership-manifest.json",
		"HARPY_BOOTSTRAP_OPERATION_PATH":                   "bootstrap-operation.json",
		"HARPY_EXTENSION_OPERATION_PATH":                   "extension-operation.json",
		"HARPY_CREDENTIAL_CLEANUP_OPERATION_PATH":          "cleanup-operation.json",
		"HARPY_CUTOVER_AUTHORITY_OPERATION_PATH":           "cutover-authority-operation.json",
		"HARPY_CUTOVER_AUTHORITY_OPERATION_SIGNATURE_PATH": "cutover-authority-operation.sig",
		"HARPY_CUTOVER_AUTHORITY_RECEIPT_PATH":             "cutover-authority-receipt.json",
		"HARPY_CUTOVER_AUTHORITY_RECEIPT_SIGNATURE_PATH":   "cutover-authority-receipt.sig",
	}
	for key, name := range paths {
		values[key] = releaseRoot + name
	}
	return values
}

func tenantMigrationExecutorEnvironment(set *artifactSet) profileEnvironment {
	values := baseRuntimeEnvironmentFor(secretRoot + "postgres-migrator-dsn.txt")
	values["HARPY_MANAGEMENT_LISTEN_ADDR"] = "0.0.0.0:9000"
	values["HARPY_RELEASE_MANIFEST_PATH"] = releaseRoot + "release.json"
	values["HARPY_RELEASE_SIGNATURE_PATH"] = releaseRoot + "release.sig"
	values["HARPY_RELEASE_TRUST_ROOT_PATH"] = releaseRoot + "trust-root.json"
	return values
}

func addTenantMigrationExecutorEnvironment(set *artifactSet) error {
	content, err := renderProfileEnvironment(tenantMigrationExecutorEnvironment(set))
	if err != nil {
		return fmt.Errorf("render tenant-migration-executor.env: %w", err)
	}
	return set.add("tenant-migration-executor.env", content)
}

func operatorEnvironment(set *artifactSet) profileEnvironment {
	values := baseRuntimeEnvironmentFor(secretRoot + "postgres-runtime-dsn.txt")
	values["HARPY_IAM_LOGIN_DIRECTORY_KEYRING_FILE"] =
		secretRoot + "iam-login-directory-keyring.json"
	values["HARPY_INSTALLATION_MANIFEST_PATH"] = "/run/installation/manifest.json"
	values["HARPY_INITIAL_ADMIN_PASSWORD_FILE"] =
		"/run/installation/initial-admin-password"
	values["HARPY_RECOVERY_SIGNING_DIRECTORY"] = "/run/recovery-signing/current"
	values["HARPY_LICENSE_SNAPSHOT_FILE"] =
		"/run/dev-authority/current/license-snapshot.json"
	values["HARPY_LICENSE_TRUST_BUNDLE_PATH"] =
		"/run/dev-authority/current/license-trust.json"
	return values
}
