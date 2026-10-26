package attestation

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

type PostgresArtifactSet struct {
	Migration          VerifiedPostgresArtifact[postgresv1.MigrationManifest]
	Extension          VerifiedPostgresArtifact[postgresv1.ExtensionManifest]
	Roles              VerifiedPostgresArtifact[postgresv1.RolesManifest]
	Privileges         VerifiedPostgresArtifact[postgresv1.PrivilegesManifest]
	Ownership          VerifiedOwnershipTree
	BootstrapOperation VerifiedPostgresArtifact[postgresv1.BootstrapOperation]
	ExtensionOperation VerifiedPostgresArtifact[postgresv1.ExtensionOperation]
	CleanupOperation   VerifiedPostgresArtifact[postgresv1.CredentialCleanupOperation]
}

type VerifiedPostgresArtifactSet struct {
	artifacts          PostgresArtifactSet
	selectedMigrations []postgresv1.SelectedMigration
	databaseOrigin     postgresv1.DatabaseOrigin
	databaseName       string
	verified           bool
}

func (set VerifiedPostgresArtifactSet) Artifacts() PostgresArtifactSet {
	return set.artifacts
}

func (set VerifiedPostgresArtifactSet) SelectedMigrations() []postgresv1.SelectedMigration {
	return append([]postgresv1.SelectedMigration(nil), set.selectedMigrations...)
}

func (set VerifiedPostgresArtifactSet) DatabaseOrigin() postgresv1.DatabaseOrigin {
	return set.databaseOrigin
}

func (set VerifiedPostgresArtifactSet) DatabaseName() string {
	return set.databaseName
}

func (set VerifiedPostgresArtifactSet) IsVerified() bool {
	return set.verified
}

func (set PostgresArtifactSet) allVerified() bool {
	return set.Migration.IsVerified() && set.Extension.IsVerified() &&
		set.Roles.IsVerified() && set.Privileges.IsVerified() && set.Ownership.IsVerified() &&
		set.BootstrapOperation.IsVerified() && set.ExtensionOperation.IsVerified() &&
		set.CleanupOperation.IsVerified()
}
