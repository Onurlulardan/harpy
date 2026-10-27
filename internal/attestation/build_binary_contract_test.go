package attestation

import "testing"

func TestBuildBinarySchemaContractBindsVerifiedReleaseArtifacts(t *testing.T) {
	release, migration, extension := binaryContractInputs(t)
	artifact, err := BuildBinarySchemaContract(release, migration, extension)
	if err != nil {
		t.Fatalf("build binary contract: %v", err)
	}
	contract := artifact.Value()
	manifest := release.Value()
	if !artifact.IsVerified() || artifact.DigestSHA256() == "" ||
		contract.RuntimeGeneration != manifest.RuntimeGeneration ||
		contract.MinimumDatabaseGeneration != manifest.MinimumDatabaseGeneration ||
		contract.MigrationManifestSHA256 != migration.DigestSHA256() ||
		contract.ExtensionManifestSHA256 != extension.DigestSHA256() {
		t.Fatalf("unexpected binary contract: %+v", contract)
	}
}

func TestBuildBinarySchemaContractRejectsUnverifiedAndMismatchedInputs(t *testing.T) {
	release, migration, extension := binaryContractInputs(t)
	if _, err := BuildBinarySchemaContract(VerifiedRelease{}, migration, extension); err == nil {
		t.Fatal("unverified release was accepted")
	}
	manifest := release.Value()
	manifest.ExtensionManifestSHA256 = manifest.MigrationManifestSHA256
	tampered := newVerifiedRelease(manifest, release.ManifestSHA256())
	if _, err := BuildBinarySchemaContract(tampered, migration, extension); err == nil {
		t.Fatal("release with mismatched extension digest was accepted")
	}
	manifest = release.Value()
	manifest.DatabaseOrigin = "v1_adopted"
	wrongOrigin := newVerifiedRelease(manifest, release.ManifestSHA256())
	if _, err := BuildBinarySchemaContract(wrongOrigin, migration, extension); err == nil {
		t.Fatal("release with mismatched extension origin was accepted")
	}
}
