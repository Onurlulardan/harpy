package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func loadExistingReleaseCatalog(
	directory string, migration artifact, contract runtimev1.BinarySchemaContract,
) (releaseCatalog, error) {
	var catalog releaseCatalog
	catalog.migration = migration
	catalog.migration.name = "migration-manifest.json"
	if err := requireExactOutputArtifact(directory, catalog.migration); err != nil {
		return releaseCatalog{}, err
	}
	var err error
	catalog.extension, err = exactExtensionArtifact()
	if err != nil {
		return releaseCatalog{}, err
	}
	catalog.roles, err = newArtifact("roles-manifest.json", rolesManifest())
	if err != nil {
		return releaseCatalog{}, err
	}
	catalog.privileges, err = newArtifact("privileges-manifest.json", privilegesManifest())
	if err != nil {
		return releaseCatalog{}, err
	}
	for _, item := range []artifact{catalog.extension, catalog.roles, catalog.privileges} {
		if err := requireExactOutputArtifact(directory, item); err != nil {
			return releaseCatalog{}, err
		}
	}
	catalog.ownership, err = loadOwnershipArtifact(directory)
	if err != nil {
		return releaseCatalog{}, err
	}
	catalog.bootstrapOperation, err = newArtifact("bootstrap-operation.json",
		bootstrapOperation(catalog.roles.digest, catalog.privileges.digest, catalog.ownership.digest))
	if err != nil {
		return releaseCatalog{}, err
	}
	catalog.extensionOperation, err = newArtifact("extension-operation.json",
		extensionOperation(catalog.extension.digest, catalog.ownership.digest))
	if err != nil {
		return releaseCatalog{}, err
	}
	catalog.cleanupOperation, err = newArtifact("cleanup-operation.json",
		cleanupOperation(catalog.roles.digest, catalog.privileges.digest))
	if err != nil {
		return releaseCatalog{}, err
	}
	catalog.binaryContract, err = newArtifact("binary-schema-contract.json", contract)
	if err != nil {
		return releaseCatalog{}, err
	}
	catalog.catalogPackageSet, err = loadCatalogPackageSetArtifact(directory)
	if err != nil {
		return releaseCatalog{}, err
	}
	catalog.schemaContract = contract
	for _, item := range []artifact{
		catalog.bootstrapOperation, catalog.extensionOperation,
		catalog.cleanupOperation, catalog.binaryContract,
	} {
		if err := requireExactOutputArtifact(directory, item); err != nil {
			return releaseCatalog{}, err
		}
	}
	return catalog, verifyPostgresCatalog(catalog)
}

func loadCatalogPackageSetArtifact(directory string) (artifact, error) {
	const name = "catalog-package-set.json"
	document, err := os.ReadFile(filepath.Join(directory, name))
	if err != nil {
		return artifact{}, fmt.Errorf("read %s: %w", name, err)
	}
	if len(document) > catalogpackagev1.MaximumPackageSetBytes {
		return artifact{}, fmt.Errorf(
			"%s exceeds the %d byte contract limit",
			name, catalogpackagev1.MaximumPackageSetBytes,
		)
	}
	if _, err := catalogpackagev1.DecodeCanonical(document); err != nil {
		return artifact{}, fmt.Errorf("verify %s: %w", name, err)
	}
	digest := sha256.Sum256(document)
	return artifact{
		name: name, digest: hex.EncodeToString(digest[:]), document: document,
	}, nil
}

func loadOwnershipArtifact(directory string) (artifact, error) {
	name := "ownership-manifest.json"
	document, err := os.ReadFile(filepath.Join(directory, name))
	if err != nil {
		return artifact{}, fmt.Errorf("read %s: %w", name, err)
	}
	verified, err := attestation.VerifyOwnershipManifestArtifact(document)
	if err != nil {
		return artifact{}, fmt.Errorf("verify %s: %w", name, err)
	}
	canonical := verified.CanonicalJSON()
	if !bytes.Equal(document, canonical) {
		return artifact{}, fmt.Errorf("%s is not canonical", name)
	}
	return artifact{name: name, digest: verified.DigestSHA256(), document: canonical}, nil
}

func requireExactOutputArtifact(directory string, expected artifact) error {
	document, err := os.ReadFile(filepath.Join(directory, expected.name))
	if err != nil {
		return fmt.Errorf("read %s: %w", expected.name, err)
	}
	if !bytes.Equal(document, expected.document) {
		return fmt.Errorf("existing %s differs from current build contract", expected.name)
	}
	return nil
}
