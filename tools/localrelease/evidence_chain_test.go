package main

import (
	"bytes"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestGeneratedEvidencePassesProductionVerifiers(t *testing.T) {
	migration, err := loadMigrationArtifact()
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := buildReleaseCatalog(migration)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	ids := newReleaseIDs()
	packageAuthority := testCatalogPackageAuthority()
	catalog.catalogPackageSet, err = buildCatalogPackageSet(
		ids.release, packageAuthority,
	)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := newSigningAuthority(now)
	if err != nil {
		t.Fatal(err)
	}
	live := exactTestDatabaseIdentity()
	authority, err := buildBootstrapAuthority(now, ids, catalog, signer)
	if err != nil {
		t.Fatal(err)
	}
	operationArtifacts, operation, err := buildCutoverOperation(
		now, ids, live, catalog, authority, signer,
	)
	if err != nil {
		t.Fatal(err)
	}
	release, releaseManifest, err := buildReleaseManifest(
		now, ids, catalog, authority, operationArtifacts, signer,
	)
	if err != nil {
		t.Fatal(err)
	}
	if releaseManifest.RuntimeGeneration != catalog.schemaContract.RuntimeGeneration ||
		releaseManifest.MinimumDatabaseGeneration !=
			catalog.schemaContract.MinimumDatabaseGeneration {
		t.Fatal("release manifest generations differ from the binary contract")
	}
	evidence, err := buildAllReceipts(
		now, ids, live, catalog, authority, operationArtifacts, operation, release, signer,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyGeneratedEvidence(
		now, catalog, signer, packageAuthority, evidence,
	); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "release")
	if err := publishGeneratedEvidence(target, catalog, signer, evidence, false); err != nil {
		t.Fatal(err)
	}
	loadedCatalog, err := loadExistingReleaseCatalog(
		target, catalog.migration, catalog.schemaContract,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := loadExistingEvidence(
		target, now, loadedCatalog, packageAuthority,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "release.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadExistingEvidence(
		target, now, loadedCatalog, packageAuthority,
	); err == nil {
		t.Fatal("corrupt existing release evidence was accepted")
	}
}

func testCatalogPackageAuthority() catalogPackageAuthority {
	privateKey := ed25519.NewKeyFromSeed(
		bytes.Repeat([]byte{0x35}, ed25519.SeedSize),
	)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	return catalogPackageAuthority{
		keyID: "catalog-test-2026-07", privateKey: privateKey,
		keys: map[string]ed25519.PublicKey{
			"catalog-test-2026-07": publicKey,
		},
	}
}

func exactTestDatabaseIdentity() liveDatabaseIdentity {
	return liveDatabaseIdentity{
		name: databaseName, owner: runtimev1.DatabaseOwnerRole,
		encoding: runtimev1.PostgresEncoding, provider: runtimev1.PostgresLocaleProvider,
		collation: runtimev1.PostgresCollation, ctype: runtimev1.PostgresCharacterType,
		systemID: "7623456789012345678", databaseOID: 16384,
		checksums:     runtimev1.PostgresDataChecksums,
		serverVersion: runtimev1.PostgresServerVersionNumber,
	}
}
