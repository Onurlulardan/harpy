package main

import (
	"fmt"
	"path/filepath"
)

func releaseCatalog() []releaseFile {
	return []releaseFile{
		{key: "HARPY_RELEASE_MANIFEST_PATH", name: "release-manifest.json"},
		{key: "HARPY_RELEASE_SIGNATURE_PATH", name: "release-signature.json"},
		{key: "HARPY_RELEASE_TRUST_ROOT_PATH", name: "trust-root.json"},
		{key: "HARPY_BOOTSTRAP_RECEIPT_PATH", name: "bootstrap-receipt.json"},
		{key: "HARPY_EXTENSION_RECEIPT_PATH", name: "extension-receipt.json"},
		{key: "HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH", name: "authority-manifest.json"},
		{key: "HARPY_BOOTSTRAP_AUTHORITY_SIGNATURE_PATH", name: "authority-signature.json"},
		{key: "HARPY_CREDENTIAL_CLEANUP_RECEIPT_PATH", name: "cleanup-receipt.json"},
		{key: "HARPY_MIGRATION_MANIFEST_PATH", name: "migration-manifest.json"},
		{key: "HARPY_EXTENSION_MANIFEST_PATH", name: "extension-manifest.json"},
		{key: "HARPY_ROLES_MANIFEST_PATH", name: "roles-manifest.json"},
		{key: "HARPY_PRIVILEGES_MANIFEST_PATH", name: "privileges-manifest.json"},
		{key: "HARPY_OWNERSHIP_MANIFEST_PATH", name: "ownership-manifest.json"},
		{key: "HARPY_BOOTSTRAP_OPERATION_PATH", name: "bootstrap-operation.json"},
		{key: "HARPY_EXTENSION_OPERATION_PATH", name: "extension-operation.json"},
		{key: "HARPY_CREDENTIAL_CLEANUP_OPERATION_PATH", name: "cleanup-operation.json"},
		{key: "HARPY_CUTOVER_AUTHORITY_OPERATION_PATH", name: "cutover-operation.json"},
		{key: "HARPY_CUTOVER_AUTHORITY_OPERATION_SIGNATURE_PATH", name: "cutover-operation-signature.json"},
		{key: "HARPY_CUTOVER_AUTHORITY_RECEIPT_PATH", name: "cutover-receipt.json"},
		{key: "HARPY_CUTOVER_AUTHORITY_RECEIPT_SIGNATURE_PATH", name: "cutover-receipt-signature.json"},
		{key: "HARPY_CATALOG_PACKAGE_SET_PATH", name: "catalog-package-set.json"},
		{key: "HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING_FILE", name: "catalog-package-keyring.json"},
	}
}

func bindReleaseBodies(bodies ...[]byte) ([]releaseFile, error) {
	catalog := releaseCatalog()
	if len(bodies) != len(catalog) {
		return nil, fmt.Errorf("release catalog does not match generated files")
	}
	files := make([]releaseFile, len(catalog))
	for index := range catalog {
		files[index] = catalog[index]
		files[index].body = bodies[index]
	}
	return files, nil
}

func reuseCompleteRelease(directory string) ([]string, bool, error) {
	catalog := releaseCatalog()
	lines := make([]string, 0, len(catalog))
	for _, file := range catalog {
		path := filepath.Join(directory, file.name)
		exists, err := installedFile(path)
		if err != nil {
			return nil, false, err
		}
		if !exists {
			return nil, false, nil
		}
		lines = append(lines, file.key+"="+path)
	}
	return lines, true, nil
}
