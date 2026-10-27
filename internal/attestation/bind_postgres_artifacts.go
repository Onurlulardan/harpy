package attestation

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func BindPostgresArtifactSet(
	release VerifiedRelease,
	authority VerifiedBootstrapAuthority,
	artifacts PostgresArtifactSet,
) (VerifiedPostgresArtifactSet, error) {
	if !authority.verified || !artifacts.allVerified() {
		return VerifiedPostgresArtifactSet{}, fmt.Errorf("PostgreSQL artifact evidence is not verified")
	}
	releaseManifest := release.Value()
	origin := postgresv1.DatabaseOrigin(releaseManifest.DatabaseOrigin)
	if !origin.Valid() || release.ManifestSHA256() == "" {
		return VerifiedPostgresArtifactSet{}, fmt.Errorf("release artifact identity is invalid")
	}
	if err := bindReleaseArtifactHashes(releaseManifest, artifacts); err != nil {
		return VerifiedPostgresArtifactSet{}, err
	}
	if err := bindAuthorityArtifacts(release, authority, artifacts); err != nil {
		return VerifiedPostgresArtifactSet{}, err
	}
	if err := bindArtifactIdentity(releaseManifest, artifacts); err != nil {
		return VerifiedPostgresArtifactSet{}, err
	}
	if err := bindOperationInputs(artifacts); err != nil {
		return VerifiedPostgresArtifactSet{}, err
	}
	migrationManifest := artifacts.Migration.Value()
	if err := requireFreshMigrationVariant(migrationManifest); err != nil {
		return VerifiedPostgresArtifactSet{}, err
	}
	selected, err := migrationManifest.Select(origin)
	if err != nil {
		return VerifiedPostgresArtifactSet{}, fmt.Errorf("select migration variants: %w", err)
	}
	return VerifiedPostgresArtifactSet{
		artifacts: artifacts, selectedMigrations: selected,
		databaseOrigin: origin, databaseName: releaseManifest.DatabaseName, verified: true,
	}, nil
}

func requireFreshMigrationVariant(manifest postgresv1.MigrationManifest) error {
	for _, entry := range manifest.Entries {
		if entry.Variants.Fresh == nil {
			return fmt.Errorf("release migration entry lacks a fresh variant")
		}
	}
	return nil
}

func bindArtifactIdentity(release ReleaseManifest, artifacts PostgresArtifactSet) error {
	origin := release.DatabaseOrigin
	database := release.DatabaseName
	extension := artifacts.Extension.Value()
	roles := artifacts.Roles.Value()
	privileges := artifacts.Privileges.Value()
	ownership := artifacts.Ownership.Root().Value()
	bootstrap := artifacts.BootstrapOperation.Value()
	extensionOperation := artifacts.ExtensionOperation.Value()
	cleanup := artifacts.CleanupOperation.Value()
	if extension.DatabaseOrigin != origin || roles.DatabaseOrigin != origin ||
		privileges.DatabaseOrigin != origin || ownership.DatabaseOrigin != origin ||
		bootstrap.DatabaseOrigin != origin || extensionOperation.DatabaseOrigin != origin ||
		cleanup.DatabaseOrigin != origin {
		return fmt.Errorf("artifact database origin differs from release")
	}
	if privileges.DatabaseName != database || ownership.DatabaseName != database ||
		bootstrap.DatabaseName != database || extensionOperation.DatabaseName != database ||
		cleanup.DatabaseName != database {
		return fmt.Errorf("artifact database name differs from release")
	}
	return nil
}
