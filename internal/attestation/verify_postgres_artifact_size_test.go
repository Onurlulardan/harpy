package attestation

import (
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestPostgresArtifactSizeBoundaries(t *testing.T) {
	limits := []int{
		runtimev1.MaximumAttestationDocumentBytes,
		postgresv1.MaximumMigrationManifestBytes,
		postgresv1.MaximumExtensionManifestBytes,
		postgresv1.MaximumRolesManifestBytes,
		postgresv1.MaximumPrivilegesManifestBytes,
		postgresv1.MaximumOwnershipManifestBytes,
		postgresv1.MaximumOwnershipChunkBytes,
		postgresv1.MaximumBootstrapOperationBytes,
		postgresv1.MaximumExtensionOperationBytes,
		postgresv1.MaximumCredentialCleanupOperationBytes,
	}
	for _, limit := range limits {
		if !withinDocumentSizeLimit(limit, limit) {
			t.Errorf("boundary %d was rejected", limit)
		}
		if withinDocumentSizeLimit(limit+1, limit) {
			t.Errorf("oversize %d was accepted", limit+1)
		}
	}
}

func TestPostgresArtifactVerifiersBindTypeSpecificLimits(t *testing.T) {
	migration, err := VerifyMigrationManifestArtifact(testDocument(t, testMigrationManifest()))
	assertArtifactLimit(t, migration, postgresv1.MaximumMigrationManifestBytes, err)
	extension, err := VerifyExtensionManifestArtifact(testDocument(t, testExtensionManifest()))
	assertArtifactLimit(t, extension, postgresv1.MaximumExtensionManifestBytes, err)
	roles, err := VerifyRolesManifestArtifact(testDocument(t, testRolesManifest()))
	assertArtifactLimit(t, roles, postgresv1.MaximumRolesManifestBytes, err)
	privileges, err := VerifyPrivilegesManifestArtifact(testDocument(t, testPrivilegesManifest()))
	assertArtifactLimit(t, privileges, postgresv1.MaximumPrivilegesManifestBytes, err)
	ownership, err := VerifyOwnershipManifestArtifact(testDocument(t, testOwnershipManifest()))
	assertArtifactLimit(t, ownership, postgresv1.MaximumOwnershipManifestBytes, err)

	hash := testHash("a")
	bootstrap, err := VerifyBootstrapOperationArtifact(testDocument(t,
		testBootstrapOperation(hash, hash, hash)))
	assertArtifactLimit(t, bootstrap, postgresv1.MaximumBootstrapOperationBytes, err)
	extensionOperation, err := VerifyExtensionOperationArtifact(testDocument(t,
		testExtensionOperation(hash, hash)))
	assertArtifactLimit(t, extensionOperation, postgresv1.MaximumExtensionOperationBytes, err)
	cleanup, err := VerifyCredentialCleanupOperationArtifact(testDocument(t,
		testCleanupOperation(hash, hash)))
	assertArtifactLimit(t, cleanup, postgresv1.MaximumCredentialCleanupOperationBytes, err)
}

func assertArtifactLimit[T postgresArtifactContract](
	t *testing.T,
	artifact VerifiedPostgresArtifact[T],
	want int,
	err error,
) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if artifact.maximumBytes != want {
		t.Fatalf("artifact limit = %d, want %d", artifact.maximumBytes, want)
	}
}
