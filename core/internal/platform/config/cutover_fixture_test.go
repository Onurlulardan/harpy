package config

import (
	"path/filepath"
	"testing"
)

func cutoverEnvironment(t *testing.T) []string {
	t.Helper()
	directory := t.TempDir()
	return []string{
		"HARPY_ENVIRONMENT=development",
		"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_CUTOVER_SOURCE_DSN=postgresql://harpy_cutover_decoder:secret@source.invalid/app?sslmode=disable",
		"HARPY_CUTOVER_TARGET_DSN=postgresql://harpy_adopter:secret@target.invalid/app?sslmode=disable",
		"HARPY_CUTOVER_LEASE_MANIFEST_PATH=" + filepath.Join(directory, "lease.json"),
		"HARPY_CUTOVER_LEASE_SIGNATURE_PATH=" + filepath.Join(directory, "lease.sig"),
		"HARPY_CUTOVER_TRUST_ROOT_PATH=" + filepath.Join(directory, "trust.json"),
		"HARPY_CUTOVER_SESSION_SEAL_RECEIPT_PATH=" + filepath.Join(directory, "seal.json"),
		"HARPY_CUTOVER_BATCH_PATH=" + filepath.Join(directory, "batch.json"),
	}
}

func cutoverFinalizationEnvironment(t *testing.T) []string {
	t.Helper()
	directory := t.TempDir()
	return []string{
		"HARPY_ENVIRONMENT=development",
		"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_CUTOVER_FINALIZER_DSN=postgresql://harpy_cutover_finalizer:secret@target.invalid/app?sslmode=disable",
		"HARPY_CUTOVER_LEASE_MANIFEST_PATH=" + filepath.Join(directory, "lease.json"),
		"HARPY_CUTOVER_LEASE_SIGNATURE_PATH=" + filepath.Join(directory, "lease.sig"),
		"HARPY_CUTOVER_TRUST_ROOT_PATH=" + filepath.Join(directory, "trust.json"),
		"HARPY_CUTOVER_SESSION_SEAL_RECEIPT_PATH=" + filepath.Join(directory, "seal.json"),
		"HARPY_CUTOVER_CREDENTIAL_CLEANUP_RECEIPT_PATH=" + filepath.Join(directory, "cleanup.json"),
		"HARPY_CUTOVER_FINALIZER_LEASE_MANIFEST_PATH=" + filepath.Join(directory, "finalizer.json"),
		"HARPY_CUTOVER_FINALIZER_LEASE_SIGNATURE_PATH=" + filepath.Join(directory, "finalizer.sig"),
		"HARPY_RELEASE_MANIFEST_PATH=" + filepath.Join(directory, "release.json"),
		"HARPY_RELEASE_SIGNATURE_PATH=" + filepath.Join(directory, "release.sig"),
		"HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH=" + filepath.Join(directory, "authority.json"),
		"HARPY_BOOTSTRAP_AUTHORITY_SIGNATURE_PATH=" + filepath.Join(directory, "authority.sig"),
		"HARPY_CUTOVER_AUTHORITY_OPERATION_PATH=" + filepath.Join(directory, "cutover-authority-operation.json"),
		"HARPY_CUTOVER_AUTHORITY_OPERATION_SIGNATURE_PATH=" + filepath.Join(directory, "cutover-authority-operation.sig"),
		"HARPY_CUTOVER_AUTHORITY_RECEIPT_PATH=" + filepath.Join(directory, "cutover-authority-receipt.json"),
		"HARPY_CUTOVER_AUTHORITY_RECEIPT_SIGNATURE_PATH=" + filepath.Join(directory, "cutover-authority-receipt.sig"),
		"HARPY_CUTOVER_TERMINAL_EVIDENCE_PATH=" + filepath.Join(directory, "terminal-evidence.json"),
	}
}
