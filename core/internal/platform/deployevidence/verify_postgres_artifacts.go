package deployevidence

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func verifyPostgresArtifacts(
	settings config.MigrationArtifacts,
) (attestation.PostgresArtifactSet, error) {
	migrationDocument, err := readArtifact(
		settings.MigrationManifestPath().Value(), postgresv1.MaximumMigrationManifestBytes,
		"migration descriptor",
	)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	migration, err := attestation.VerifyMigrationManifestArtifact(migrationDocument.bytes)
	if err != nil {
		return attestation.PostgresArtifactSet{}, fmt.Errorf("migration descriptor verification failed")
	}
	extensionDocument, err := readArtifact(
		settings.ExtensionManifestPath().Value(), postgresv1.MaximumExtensionManifestBytes,
		"extension descriptor",
	)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	extension, err := attestation.VerifyExtensionManifestArtifact(extensionDocument.bytes)
	if err != nil {
		return attestation.PostgresArtifactSet{}, fmt.Errorf("extension descriptor verification failed")
	}
	rolesDocument, err := readArtifact(
		settings.RolesManifestPath().Value(), postgresv1.MaximumRolesManifestBytes,
		"roles descriptor",
	)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	roles, err := attestation.VerifyRolesManifestArtifact(rolesDocument.bytes)
	if err != nil {
		return attestation.PostgresArtifactSet{}, fmt.Errorf("roles descriptor verification failed")
	}
	privilegesDocument, err := readArtifact(
		settings.PrivilegesManifestPath().Value(), postgresv1.MaximumPrivilegesManifestBytes,
		"privileges descriptor",
	)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	privileges, err := attestation.VerifyPrivilegesManifestArtifact(privilegesDocument.bytes)
	if err != nil {
		return attestation.PostgresArtifactSet{}, fmt.Errorf("privileges descriptor verification failed")
	}
	ownership, err := loadOwnershipTree(settings.OwnershipManifestPath().Value())
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	operations, err := verifyPostgresOperations(settings)
	if err != nil {
		return attestation.PostgresArtifactSet{}, err
	}
	return attestation.PostgresArtifactSet{
		Migration: migration, Extension: extension, Roles: roles, Privileges: privileges,
		Ownership: ownership, BootstrapOperation: operations.bootstrap,
		ExtensionOperation: operations.extension, CleanupOperation: operations.cleanup,
	}, nil
}
