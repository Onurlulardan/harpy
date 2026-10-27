package attestation

import (
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestBindPostgresArtifactSetRejectsMissingFreshVariant(t *testing.T) {
	release, authority, artifacts := testVerifiedArtifactSet(t)
	manifest := artifacts.Migration.Value()
	manifest.Entries[0].Variants.Fresh = nil
	bound, err := bindMigrationManifest(t, release, authority, artifacts, manifest)
	if err == nil || bound.IsVerified() {
		t.Fatal("release without a fresh migration variant was accepted")
	}
}

func TestBindPostgresArtifactSetAcceptsFreshOnlyVariant(t *testing.T) {
	release, authority, artifacts := testVerifiedArtifactSet(t)
	manifest := artifacts.Migration.Value()
	manifest.Entries[0].Variants.V1Adopted = nil
	bound, err := bindMigrationManifest(t, release, authority, artifacts, manifest)
	if err != nil {
		t.Fatal(err)
	}
	selected := bound.SelectedMigrations()
	if !bound.IsVerified() || len(selected) != 1 || selected[0].Origin != "fresh" ||
		selected[0].Path != "migrations/20260714010000_foundation.fresh.sql" {
		t.Fatal("fresh-only migration was not selected")
	}
}

func bindMigrationManifest(
	t *testing.T,
	release VerifiedRelease,
	authority VerifiedBootstrapAuthority,
	artifacts PostgresArtifactSet,
	manifest postgresv1.MigrationManifest,
) (VerifiedPostgresArtifactSet, error) {
	t.Helper()
	verified, err := VerifyMigrationManifestArtifact(testDocument(t, manifest))
	if err != nil {
		t.Fatal(err)
	}
	artifacts.Migration = verified
	value := release.Value()
	value.MigrationManifestSHA256 = verified.DigestSHA256()
	return BindPostgresArtifactSet(
		newVerifiedRelease(value, release.ManifestSHA256()), authority, artifacts,
	)
}
