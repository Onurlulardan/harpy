package attestation

import (
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func binaryContractInputs(
	t *testing.T,
) (
	VerifiedRelease,
	VerifiedPostgresArtifact[postgresv1.MigrationManifest],
	VerifiedPostgresArtifact[postgresv1.ExtensionManifest],
) {
	t.Helper()
	migration, err := VerifyMigrationManifestArtifact(testDocument(t, testMigrationManifest()))
	if err != nil {
		t.Fatalf("verify migration fixture: %v", err)
	}
	extension, err := VerifyExtensionManifestArtifact(testDocument(t, testExtensionManifest()))
	if err != nil {
		t.Fatalf("verify extension fixture: %v", err)
	}
	manifestValue := validReleaseTestManifest()
	manifestValue["database_origin"] = "fresh"
	manifestValue["migration_manifest_sha256"] = migration.DigestSHA256()
	manifestValue["extension_manifest_sha256"] = extension.DigestSHA256()
	document := mustReleaseJSON(t, manifestValue)
	var manifest ReleaseManifest
	if err := decodeStrict(document, &manifest); err != nil {
		t.Fatalf("decode release fixture: %v", err)
	}
	digest, err := DigestHex(document)
	if err != nil {
		t.Fatalf("digest release fixture: %v", err)
	}
	return newVerifiedRelease(manifest, digest), migration, extension
}
