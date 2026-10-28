package main

import "testing"

func TestMigratorEnvironmentUsesLocalReleaseArtifactNames(t *testing.T) {
	if releaseRoot != "/run/release/current/" {
		t.Fatalf("release root = %q", releaseRoot)
	}
	values := migratorEnvironment(newArtifactSet())
	want := map[string]string{
		"HARPY_RELEASE_MANIFEST_PATH":                      "release.json",
		"HARPY_RELEASE_SIGNATURE_PATH":                     "release.sig",
		"HARPY_RELEASE_TRUST_ROOT_PATH":                    "trust-root.json",
		"HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH":          "bootstrap-authority.json",
		"HARPY_BOOTSTRAP_AUTHORITY_SIGNATURE_PATH":         "bootstrap-authority.sig",
		"HARPY_BOOTSTRAP_RECEIPT_PATH":                     "bootstrap-receipt.json",
		"HARPY_EXTENSION_RECEIPT_PATH":                     "extension-receipt.json",
		"HARPY_CREDENTIAL_CLEANUP_RECEIPT_PATH":            "cleanup-receipt.json",
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
	for key, name := range want {
		if values[key] != releaseRoot+name {
			t.Fatalf("%s=%q, want %q", key, values[key], releaseRoot+name)
		}
	}
}
