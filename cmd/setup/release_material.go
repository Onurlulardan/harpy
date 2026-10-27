package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type releaseFile struct {
	key  string
	name string
	body []byte
}

func releaseMaterial(root, profile string) ([]string, error) {
	if err := validProfile(profile); err != nil {
		return nil, err
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	directory := filepath.Join(absolute, "deploy", profile, "secrets", "release")
	if err = os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	files, err := buildReleaseFiles(newReleaseClock(time.Now()))
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, len(files))
	for _, file := range files {
		path := filepath.Join(directory, file.name)
		if err = writePrivate(path, file.body); err != nil {
			return nil, fmt.Errorf("write %s: %w", file.name, err)
		}
		lines = append(lines, file.key+"="+path)
	}
	return lines, nil
}

func buildReleaseFiles(clock releaseClock) ([]releaseFile, error) {
	keys, err := newReleaseKeys(clock)
	if err != nil {
		return nil, err
	}
	manifests, err := buildPostgresManifests()
	if err != nil {
		return nil, fmt.Errorf("postgres manifests: %w", err)
	}
	operations, err := buildPostgresOperations(manifests)
	if err != nil {
		return nil, fmt.Errorf("postgres operations: %w", err)
	}
	authority, authoritySignature, err := buildAuthority(keys, manifests, operations)
	if err != nil {
		return nil, fmt.Errorf("bootstrap authority: %w", err)
	}
	verifiedAuthority, err := attestation.VerifyBootstrapAuthority(
		authority.body, authoritySignature, keys.root, clock.now,
	)
	if err != nil {
		return nil, err
	}
	catalog, err := buildCatalogSet(keys)
	if err != nil {
		return nil, fmt.Errorf("catalog package set: %w", err)
	}
	operationDocument, operationSignature, err := buildCutoverOperation(
		keys, manifests.migration.digest, authority.digest,
	)
	if err != nil {
		return nil, fmt.Errorf("cutover authority operation: %w", err)
	}
	releaseDocument, releaseSignature, verifiedRelease, err := buildReleaseManifest(
		keys, manifests, operations, authority.digest, catalogSetDigest(catalog), operationDocument.digest,
	)
	if err != nil {
		return nil, fmt.Errorf("release manifest: %w", err)
	}
	if err = verifyCatalogSet(keys, catalog, verifiedRelease); err != nil {
		return nil, fmt.Errorf("catalog package set: %w", err)
	}
	verifiedOperation, err := verifyCutoverOperation(
		keys, operationDocument.body, operationSignature, verifiedRelease, verifiedAuthority,
	)
	if err != nil {
		return nil, fmt.Errorf("cutover authority operation: %w", err)
	}
	var operation cutoverv1.AuthorityAdoptionOperation
	if err = attestation.DecodeStrict(operationDocument.body, &operation); err != nil {
		return nil, err
	}
	receipts, err := localRuntimeReceipts(keys, manifests, operations, authority, releaseDocument)
	if err != nil {
		return nil, fmt.Errorf("runtime receipts: %w", err)
	}
	cutoverReceipt, cutoverReceiptSignature, err := buildCutoverReceipt(
		keys, operation, verifiedOperation, verifiedRelease, verifiedAuthority,
	)
	if err != nil {
		return nil, fmt.Errorf("cutover authority receipt: %w", err)
	}
	return []releaseFile{
		{"HARPY_RELEASE_MANIFEST_PATH", "release-manifest.json", releaseDocument.body},
		{"HARPY_RELEASE_SIGNATURE_PATH", "release-signature.json", releaseSignature},
		{"HARPY_RELEASE_TRUST_ROOT_PATH", "trust-root.json", keys.rootDocument},
		{"HARPY_BOOTSTRAP_RECEIPT_PATH", "bootstrap-receipt.json", receipts[0]},
		{"HARPY_EXTENSION_RECEIPT_PATH", "extension-receipt.json", receipts[1]},
		{"HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH", "authority-manifest.json", authority.body},
		{"HARPY_BOOTSTRAP_AUTHORITY_SIGNATURE_PATH", "authority-signature.json", authoritySignature},
		{"HARPY_CREDENTIAL_CLEANUP_RECEIPT_PATH", "cleanup-receipt.json", receipts[2]},
		{"HARPY_MIGRATION_MANIFEST_PATH", "migration-manifest.json", manifests.migration.body},
		{"HARPY_EXTENSION_MANIFEST_PATH", "extension-manifest.json", manifests.extension.body},
		{"HARPY_ROLES_MANIFEST_PATH", "roles-manifest.json", manifests.roles.body},
		{"HARPY_PRIVILEGES_MANIFEST_PATH", "privileges-manifest.json", manifests.privileges.body},
		{"HARPY_OWNERSHIP_MANIFEST_PATH", "ownership-manifest.json", manifests.ownership.body},
		{"HARPY_BOOTSTRAP_OPERATION_PATH", "bootstrap-operation.json", operations.bootstrap.body},
		{"HARPY_EXTENSION_OPERATION_PATH", "extension-operation.json", operations.extension.body},
		{"HARPY_CREDENTIAL_CLEANUP_OPERATION_PATH", "cleanup-operation.json", operations.cleanup.body},
		{"HARPY_CUTOVER_AUTHORITY_OPERATION_PATH", "cutover-operation.json", operationDocument.body},
		{"HARPY_CUTOVER_AUTHORITY_OPERATION_SIGNATURE_PATH", "cutover-operation-signature.json", operationSignature},
		{"HARPY_CUTOVER_AUTHORITY_RECEIPT_PATH", "cutover-receipt.json", cutoverReceipt},
		{"HARPY_CUTOVER_AUTHORITY_RECEIPT_SIGNATURE_PATH", "cutover-receipt-signature.json", cutoverReceiptSignature},
		{"HARPY_CATALOG_PACKAGE_SET_PATH", "catalog-package-set.json", catalog},
		{"HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING_FILE", "catalog-package-keyring.json", keys.keyring},
	}, nil
}

func localRuntimeReceipts(
	keys releaseKeys, manifests postgresManifests, operations postgresOperations,
	authority, release releaseDocument,
) ([][]byte, error) {
	bootstrap, extension, cleanup, err := buildRuntimeReceipts(
		keys, manifests, operations, authority, release,
	)
	if err != nil {
		return nil, err
	}
	return [][]byte{bootstrap, extension, cleanup}, nil
}
