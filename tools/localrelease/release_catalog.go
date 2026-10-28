package main

import (
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/google/uuid"
)

type releaseCatalog struct {
	migration, extension, roles, privileges, ownership       artifact
	bootstrapOperation, extensionOperation, cleanupOperation artifact
	binaryContract                                           artifact
	catalogPackageSet                                        artifact
	schemaContract                                           runtimev1.BinarySchemaContract
}

func buildReleaseCatalog(migration artifact) (releaseCatalog, error) {
	var catalog releaseCatalog
	var err error
	catalog.migration = migration
	catalog.migration.name = "migration-manifest.json"
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
	catalog.ownership, err = newArtifact(
		"ownership-manifest.json", ownershipManifest(uuid.NewString()),
	)
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
	catalog.schemaContract, err = localSchemaContract(
		catalog.migration, catalog.extension.digest,
	)
	if err != nil {
		return releaseCatalog{}, err
	}
	catalog.binaryContract, err = newArtifact(
		"binary-schema-contract.json", catalog.schemaContract,
	)
	if err != nil {
		return releaseCatalog{}, err
	}
	return catalog, verifyPostgresCatalog(catalog)
}

func verifyPostgresCatalog(catalog releaseCatalog) error {
	set, err := verifiedPostgresArtifacts(catalog)
	if err != nil {
		return err
	}
	if !set.Migration.IsVerified() || !set.Ownership.IsVerified() {
		return fmt.Errorf("PostgreSQL artifact verification did not complete")
	}
	return nil
}

func verifiedPostgresArtifacts(
	catalog releaseCatalog,
) (attestation.PostgresArtifactSet, error) {
	migration, err := attestation.VerifyMigrationManifestArtifact(catalog.migration.document)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	extension, err := attestation.VerifyExtensionManifestArtifact(catalog.extension.document)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	roles, err := attestation.VerifyRolesManifestArtifact(catalog.roles.document)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	privileges, err := attestation.VerifyPrivilegesManifestArtifact(catalog.privileges.document)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	root, err := attestation.VerifyOwnershipManifestArtifact(catalog.ownership.document)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	ownership, err := attestation.VerifyOwnershipTree(root, nil)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	bootstrap, err := attestation.VerifyBootstrapOperationArtifact(catalog.bootstrapOperation.document)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	extensionOperation, err := attestation.VerifyExtensionOperationArtifact(
		catalog.extensionOperation.document,
	)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	cleanup, err := attestation.VerifyCredentialCleanupOperationArtifact(
		catalog.cleanupOperation.document,
	)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	return attestation.PostgresArtifactSet{
		Migration: migration, Extension: extension, Roles: roles, Privileges: privileges,
		Ownership: ownership, BootstrapOperation: bootstrap,
		ExtensionOperation: extensionOperation, CleanupOperation: cleanup,
	}, nil
}

func (catalog releaseCatalog) baseArtifacts() []artifact {
	return []artifact{
		catalog.migration, catalog.extension, catalog.roles, catalog.privileges,
		catalog.ownership, catalog.bootstrapOperation, catalog.extensionOperation,
		catalog.cleanupOperation, catalog.binaryContract, catalog.catalogPackageSet,
	}
}
