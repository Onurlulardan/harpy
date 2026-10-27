package attestation

import (
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestBindPostgresArtifactSetClosesHashOriginAndVariantChain(t *testing.T) {
	release, authority, artifacts := testVerifiedArtifactSet(t)
	bound, err := BindPostgresArtifactSet(release, authority, artifacts)
	if err != nil {
		t.Fatal(err)
	}
	selected := bound.SelectedMigrations()
	if !bound.IsVerified() || bound.DatabaseName() != "harpy" ||
		bound.DatabaseOrigin() != "fresh" || len(selected) != 1 ||
		selected[0].Origin != "fresh" || selected[0].Path !=
		"migrations/20260714010000_foundation.fresh.sql" {
		t.Fatal("bound artifact set lost release-selected migration identity")
	}
	selected[0].Path = "mutated"
	if bound.SelectedMigrations()[0].Path == "mutated" {
		t.Fatal("bound artifact set exposes mutable migration state")
	}
}

func TestBindPostgresArtifactSetRejectsCrossDocumentTamper(t *testing.T) {
	t.Run("release digest", func(t *testing.T) {
		release, authority, artifacts := testVerifiedArtifactSet(t)
		manifest := release.Value()
		manifest.RolesManifestSHA256 = testHash("f")
		release = newVerifiedRelease(manifest, release.ManifestSHA256())
		if _, err := BindPostgresArtifactSet(release, authority, artifacts); err == nil {
			t.Fatal("release digest tamper accepted")
		}
	})

	t.Run("operation input", func(t *testing.T) {
		release, authority, artifacts := testVerifiedArtifactSet(t)
		operation := testBootstrapOperation(
			testHash("f"), artifacts.Privileges.DigestSHA256(),
			artifacts.Ownership.Root().DigestSHA256(),
		)
		verified, err := VerifyBootstrapOperationArtifact(testDocument(t, operation))
		if err != nil {
			t.Fatal(err)
		}
		artifacts.BootstrapOperation = verified
		releaseManifest := release.Value()
		releaseManifest.BootstrapOperationDigest = verified.DigestSHA256()
		release = newVerifiedRelease(releaseManifest, release.ManifestSHA256())
		authority.manifest.BootstrapOperationDigest = verified.DigestSHA256()
		if _, err := BindPostgresArtifactSet(release, authority, artifacts); err == nil {
			t.Fatal("operation input hash tamper accepted")
		}
	})

	t.Run("origin", func(t *testing.T) {
		release, authority, artifacts := testVerifiedArtifactSet(t)
		extension := testExtensionManifest()
		extension.DatabaseOrigin = "v1_adopted"
		verified, err := VerifyExtensionManifestArtifact(testDocument(t, extension))
		if err != nil {
			t.Fatal(err)
		}
		artifacts.Extension = verified
		releaseManifest := release.Value()
		releaseManifest.ExtensionManifestSHA256 = verified.DigestSHA256()
		release = newVerifiedRelease(releaseManifest, release.ManifestSHA256())
		if _, err := BindPostgresArtifactSet(release, authority, artifacts); err == nil {
			t.Fatal("cross-document origin tamper accepted")
		}
	})

	t.Run("unverified wrapper", func(t *testing.T) {
		release, authority, artifacts := testVerifiedArtifactSet(t)
		artifacts = PostgresArtifactSet{}
		if _, err := BindPostgresArtifactSet(release, authority, artifacts); err == nil {
			t.Fatal("unverified artifact wrapper accepted")
		}
	})
}

func testVerifiedArtifactSet(
	t *testing.T,
) (VerifiedRelease, VerifiedBootstrapAuthority, PostgresArtifactSet) {
	t.Helper()
	migration, err := VerifyMigrationManifestArtifact(testDocument(t, testMigrationManifest()))
	if err != nil {
		t.Fatal(err)
	}
	extension, err := VerifyExtensionManifestArtifact(testDocument(t, testExtensionManifest()))
	if err != nil {
		t.Fatal(err)
	}
	roles, err := VerifyRolesManifestArtifact(testDocument(t, testRolesManifest()))
	if err != nil {
		t.Fatal(err)
	}
	privileges, err := VerifyPrivilegesManifestArtifact(testDocument(t, testPrivilegesManifest()))
	if err != nil {
		t.Fatal(err)
	}
	ownershipRoot, err := VerifyOwnershipManifestArtifact(testDocument(t, testOwnershipManifest()))
	if err != nil {
		t.Fatal(err)
	}
	ownership, err := VerifyOwnershipTree(ownershipRoot, nil)
	if err != nil {
		t.Fatal(err)
	}
	bootstrap, err := VerifyBootstrapOperationArtifact(testDocument(t, testBootstrapOperation(
		roles.DigestSHA256(), privileges.DigestSHA256(), ownershipRoot.DigestSHA256(),
	)))
	if err != nil {
		t.Fatal(err)
	}
	extensionOperation, err := VerifyExtensionOperationArtifact(testDocument(t,
		testExtensionOperation(extension.DigestSHA256(), ownershipRoot.DigestSHA256())))
	if err != nil {
		t.Fatal(err)
	}
	cleanup, err := VerifyCredentialCleanupOperationArtifact(testDocument(t,
		testCleanupOperation(roles.DigestSHA256(), privileges.DigestSHA256())))
	if err != nil {
		t.Fatal(err)
	}
	artifacts := PostgresArtifactSet{
		Migration: migration, Extension: extension, Roles: roles, Privileges: privileges,
		Ownership: ownership, BootstrapOperation: bootstrap,
		ExtensionOperation: extensionOperation, CleanupOperation: cleanup,
	}
	releaseID := "44444444-4444-4444-8444-444444444444"
	authorityDigest := testHash("9")
	releaseManifest := ReleaseManifest{
		ReleaseID: releaseID, DatabaseOrigin: "fresh", DatabaseName: "harpy",
		MigrationManifestSHA256: migration.DigestSHA256(),
		ExtensionManifestSHA256: extension.DigestSHA256(), RolesManifestSHA256: roles.DigestSHA256(),
		PrivilegesManifestSHA256:        privileges.DigestSHA256(),
		OwnershipManifestSHA256:         ownershipRoot.DigestSHA256(),
		BootstrapAuthoritySHA256:        authorityDigest,
		BootstrapOperationDigest:        bootstrap.DigestSHA256(),
		ExtensionOperationDigest:        extensionOperation.DigestSHA256(),
		CredentialCleanupDigest:         cleanup.DigestSHA256(),
		CutoverAuthorityOperationDigest: testHash("f"),
	}
	release := newVerifiedRelease(releaseManifest, testHash("8"))
	authority := VerifiedBootstrapAuthority{
		manifest: runtimev1.BootstrapAuthorityManifest{
			ReleaseID: releaseID, DatabaseOrigin: "fresh",
			Database:                         runtimev1.BootstrapAuthorityDatabase{Name: "harpy"},
			RolesManifestSHA256:              roles.DigestSHA256(),
			PrivilegesManifestSHA256:         privileges.DigestSHA256(),
			OwnershipManifestSHA256:          ownershipRoot.DigestSHA256(),
			BootstrapOperationDigest:         bootstrap.DigestSHA256(),
			CredentialCleanupOperationDigest: cleanup.DigestSHA256(),
		},
		manifestDigest: authorityDigest, verified: true,
	}
	return release, authority, artifacts
}
