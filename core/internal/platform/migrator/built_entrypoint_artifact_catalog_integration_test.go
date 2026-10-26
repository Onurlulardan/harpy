//go:build integration

package migrator

import (
	"testing"
)

type builtEntrypointArtifactCatalog struct {
	migration                          builtEntrypointArtifact
	catalogPackageSet                  builtEntrypointArtifact
	catalogPackageKeyring              string
	extension                          builtEntrypointArtifact
	roles                              builtEntrypointArtifact
	privileges                         builtEntrypointArtifact
	ownership                          builtEntrypointArtifact
	bootstrapOperation                 builtEntrypointArtifact
	extensionOperation                 builtEntrypointArtifact
	cleanupOperation                   builtEntrypointArtifact
	cutoverAuthorityOperation          builtEntrypointArtifact
	cutoverAuthorityOperationSignature builtEntrypointArtifact
}

func prepareBuiltEntrypointArtifactCatalog(
	t *testing.T,
	directory string,
	stage builtEntrypointStage,
) builtEntrypointArtifactCatalog {
	t.Helper()
	migration := readBuiltEntrypointMigrationArtifact(t, directory, stage)
	catalogPackageSet, catalogPackageKeyring :=
		prepareBuiltEntrypointCatalogPackageSet(t, directory, stage)
	extension := writeBuiltEntrypointArtifact(
		t, directory, "extensions.json", builtEntrypointExtensionManifest(),
	)
	if extension.digest != stage.extensionDigest {
		t.Fatalf("built extension artifact digest = %s", extension.digest)
	}
	roles := writeBuiltEntrypointArtifact(
		t, directory, "roles.json", builtEntrypointRolesManifest(),
	)
	privileges := writeBuiltEntrypointArtifact(
		t, directory, "privileges.json", builtEntrypointPrivilegesManifest(),
	)
	ownership := writeBuiltEntrypointArtifact(
		t, directory, "ownership.json", builtEntrypointOwnershipManifest(),
	)
	bootstrapOperation := writeBuiltEntrypointArtifact(
		t, directory, "bootstrap-operation.json",
		builtEntrypointBootstrapOperation(roles.digest, privileges.digest, ownership.digest),
	)
	extensionOperation := writeBuiltEntrypointArtifact(
		t, directory, "extension-operation.json",
		builtEntrypointExtensionOperation(extension.digest, ownership.digest),
	)
	cleanupOperation := writeBuiltEntrypointArtifact(
		t, directory, "cleanup-operation.json",
		builtEntrypointCleanupOperation(roles.digest, privileges.digest),
	)
	return builtEntrypointArtifactCatalog{
		migration: migration, catalogPackageSet: catalogPackageSet,
		catalogPackageKeyring: catalogPackageKeyring,
		extension:             extension, roles: roles, privileges: privileges,
		ownership: ownership, bootstrapOperation: bootstrapOperation,
		extensionOperation: extensionOperation, cleanupOperation: cleanupOperation,
	}
}
