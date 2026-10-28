package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestRefreshSignedReleaseRestoresAppliedReleaseID(t *testing.T) {
	target := filepath.Join(t.TempDir(), "current")
	output, err := beginOwnedOutput(target, false)
	if err != nil {
		t.Fatalf("begin owned output: %v", err)
	}
	applied := "01a04f8f-6d9d-74a4-97fb-1312bb4e43de"
	now := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	manifest := testReleaseManifest(applied, now)
	document, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(output.staging, "release.json"), document, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := output.commit(); err != nil {
		t.Fatalf("commit fixture: %v", err)
	}
	if err := refreshSignedRelease(target, applied, now.Add(2*time.Hour)); err != nil {
		t.Fatalf("refresh signed release: %v", err)
	}
	got, err := loadReleaseManifest(target)
	if err != nil {
		t.Fatalf("load refreshed release: %v", err)
	}
	if got.ReleaseID != applied {
		t.Fatalf("release id = %s", got.ReleaseID)
	}
	expiresAt, err := time.Parse(time.RFC3339, got.ExpiresAt)
	if err != nil || !now.Add(2*time.Hour).Before(expiresAt) {
		t.Fatalf("expiry was not extended: %s", got.ExpiresAt)
	}
}

func testReleaseManifest(releaseID string, now time.Time) attestation.ReleaseManifest {
	return attestation.ReleaseManifest{
		ContractVersion: "postgres-release-manifest.v1", ReleaseID: releaseID,
		IssuedAt: now.Add(-time.Minute).Format(time.RFC3339),
		ExpiresAt: now.Add(time.Hour).Format(time.RFC3339), DatabaseName: "harpy",
		DatabaseOrigin: "fresh", DatabaseEncoding: "UTF8", DatabaseLocaleProvider: "libc",
		DatabaseCollation: "en_US.utf8", DatabaseCType: "en_US.utf8",
		DatabaseOwner: "harpy_owner", DataChecksums: true,
		PostgresImageDigest: "sha256:c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84",
		PostgresEngineVersion: "18.4", VectorExtensionVersion: "0.8.3",
		RuntimeGeneration: 24, MinimumDatabaseGeneration: 97,
		MigrationManifestSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ExtensionManifestSHA256: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		RolesManifestSHA256: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		PrivilegesManifestSHA256: "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
		OwnershipManifestSHA256: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		BootstrapAuthoritySHA256: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		BootstrapOperationDigest: "1111111111111111111111111111111111111111111111111111111111111111",
		ExtensionOperationDigest: "2222222222222222222222222222222222222222222222222222222222222222",
		CredentialCleanupDigest: "3333333333333333333333333333333333333333333333333333333333333333",
		CutoverAuthorityOperationDigest: "4444444444444444444444444444444444444444444444444444444444444444",
		CatalogPackageSetSHA256: "5555555555555555555555555555555555555555555555555555555555555555",
	}
}
