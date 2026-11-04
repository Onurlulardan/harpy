package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/google/uuid"
)

func TestDevelopmentBootstrapAuthorityMatchesCatalogDigests(t *testing.T) {
	now := time.Date(2026, 11, 4, 16, 0, 0, 0, time.UTC)
	catalog, err := developmentAuthorityCatalog()
	if err != nil {
		t.Fatal(err)
	}
	signer, err := newSigningAuthority(now)
	if err != nil {
		t.Fatal(err)
	}
	pair, err := buildBootstrapAuthority(now, newReleaseIDs(), catalog, signer)
	if err != nil {
		t.Fatal(err)
	}
	manifest := decodeAuthority(t, pair.document.document)
	if err := runtimev1.ValidateBootstrapAuthorityManifest(manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.ContractVersion != "bootstrap-authority-manifest.v1" ||
		manifest.DatabaseOrigin != "fresh" {
		t.Fatalf("authority header = %s %s", manifest.ContractVersion, manifest.DatabaseOrigin)
	}
	environmentID, err := uuid.Parse(manifest.EnvironmentID)
	if err != nil || environmentID.Version() != uuid.Version(7) {
		t.Fatalf("environment ID = %s", manifest.EnvironmentID)
	}
	if manifest.RolesManifestSHA256 != catalog.roles.digest ||
		manifest.PrivilegesManifestSHA256 != catalog.privileges.digest ||
		manifest.OwnershipManifestSHA256 != catalog.ownership.digest ||
		manifest.BootstrapOperationDigest != catalog.bootstrapOperation.digest ||
		manifest.CredentialCleanupOperationDigest != catalog.cleanupOperation.digest {
		t.Fatal("authority digests do not match the catalog artifacts")
	}
	assertNoMigratorRole(t, manifest)
	if err := runtimev1.ValidateBootstrapAuthoritySignature(decodeSignature(t, pair.signature.document)); err != nil {
		t.Fatal(err)
	}
}

func TestPublishDevelopmentBootstrapAuthorityReplacesOwnedOutput(t *testing.T) {
	now := time.Date(2026, 11, 4, 16, 0, 0, 0, time.UTC)
	directory := filepath.Join(t.TempDir(), "current")
	if developmentAuthorityDirectory != "/run/release/current" {
		t.Fatalf("container path = %s", developmentAuthorityDirectory)
	}
	if err := publishDevelopmentBootstrapAuthority(directory, now); err != nil {
		t.Fatal(err)
	}
	first := readPublishedAuthority(t, directory)
	assertPublishedLifetime(t, first, now)
	assertPublishedOmitsRetiredRoles(t, directory)
	if err := publishDevelopmentBootstrapAuthority(directory, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	second := readPublishedAuthority(t, directory)
	if first.EnvironmentID != second.EnvironmentID ||
		first.ReleaseID != second.ReleaseID ||
		first.ManifestID != second.ManifestID {
		t.Fatal("replacement changed the stored authority identity")
	}
	assertPublishedLifetime(t, second, now.Add(time.Minute))
	if err := runArtifact([]string{"extra"}); err == nil {
		t.Fatal("artifact accepted arguments")
	}
}

func readPublishedAuthority(t *testing.T, directory string) runtimev1.BootstrapAuthorityManifest {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(directory, "bootstrap-authority.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest := decodeAuthority(t, content)
	if err := runtimev1.ValidateBootstrapAuthorityManifest(manifest); err != nil {
		t.Fatal(err)
	}
	environmentID, err := uuid.Parse(manifest.EnvironmentID)
	if err != nil || environmentID.Version() != uuid.Version(7) {
		t.Fatalf("published environment ID = %s", manifest.EnvironmentID)
	}
	signature, err := os.ReadFile(filepath.Join(directory, "bootstrap-authority.sig"))
	if err != nil {
		t.Fatal(err)
	}
	if err := runtimev1.ValidateBootstrapAuthoritySignature(decodeSignature(t, signature)); err != nil {
		t.Fatal(err)
	}
	return manifest
}

func decodeAuthority(t *testing.T, document []byte) runtimev1.BootstrapAuthorityManifest {
	t.Helper()
	var manifest runtimev1.BootstrapAuthorityManifest
	if err := json.Unmarshal(document, &manifest); err != nil {
		t.Fatal(err)
	}
	return manifest
}

func decodeSignature(t *testing.T, document []byte) runtimev1.BootstrapAuthoritySignature {
	t.Helper()
	var signature runtimev1.BootstrapAuthoritySignature
	if err := json.Unmarshal(document, &signature); err != nil {
		t.Fatal(err)
	}
	return signature
}

func assertNoMigratorRole(t *testing.T, manifest runtimev1.BootstrapAuthorityManifest) {
	t.Helper()
	for _, role := range manifest.Roles {
		if role.Name == "harpy_migrator" || strings.Contains(role.Name, "tenant_migration") {
			t.Fatalf("authority role %s was added", role.Name)
		}
	}
	source, err := os.ReadFile("development_authority.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, forbidden := range []string{
		"executePrepare", "schema-executor", "tenant-migration-executor", "harpy_migrator",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("development authority references %s", forbidden)
		}
	}
}
