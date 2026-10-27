package main

import (
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestReleaseMaterialValidityWindow(t *testing.T) {
	generated := time.Date(2026, 10, 27, 21, 36, 23, 0, time.UTC)
	files, err := buildReleaseFiles(newReleaseClock(generated))
	if err != nil {
		t.Fatalf("build release files: %v", err)
	}
	body := releaseFileBody(t, files)
	var manifest attestation.ReleaseManifest
	if err = attestation.DecodeStrict(body["release-manifest.json"], &manifest); err != nil {
		t.Fatal(err)
	}
	expires := mustStamp(t, manifest.ExpiresAt)
	if !expires.Equal(generated.Add(localReleaseValidity)) || localReleaseValidity < 24*time.Hour {
		t.Fatalf("release expires at %s", expires.Format(time.RFC3339))
	}
	root, err := attestation.ParseTrustRoot(body["trust-root.json"])
	if err != nil {
		t.Fatal(err)
	}
	if _, err = attestation.VerifyRelease(
		body["release-manifest.json"], body["release-signature.json"], root, generated, localDatabase,
	); err != nil {
		t.Fatalf("release at generation: %v", err)
	}
	if _, err = attestation.VerifyRelease(
		body["release-manifest.json"], body["release-signature.json"], root,
		expires.Add(-time.Second), localDatabase,
	); err != nil {
		t.Fatalf("release inside window: %v", err)
	}
	if _, err = attestation.VerifyRelease(
		body["release-manifest.json"], body["release-signature.json"], root, expires, localDatabase,
	); err == nil {
		t.Fatal("expiry check accepted a release at expires_at")
	}
	if _, err = attestation.VerifyBootstrapAuthority(
		body["authority-manifest.json"], body["authority-signature.json"], root,
		generated.Add(23*time.Hour),
	); err != nil {
		t.Fatalf("authority inside release window: %v", err)
	}
	var authority runtimev1.BootstrapAuthorityManifest
	if err = attestation.DecodeStrict(body["authority-manifest.json"], &authority); err != nil {
		t.Fatal(err)
	}
	for index, lease := range authority.CredentialLeases {
		span := mustStamp(t, lease.ExpiresAt).Sub(mustStamp(t, lease.NotBefore))
		if span <= 0 || span > 15*time.Minute {
			t.Fatalf("lease %d validity is %s", index, span)
		}
	}
	var operation cutoverv1.AuthorityAdoptionOperation
	if err = attestation.DecodeStrict(body["cutover-operation.json"], &operation); err != nil {
		t.Fatal(err)
	}
	span := mustStamp(t, operation.ExpiresAt).Sub(mustStamp(t, operation.IssuedAt))
	if span <= 0 || span > 15*time.Minute {
		t.Fatalf("cutover operation validity is %s", span)
	}
}

func releaseFileBody(t *testing.T, files []releaseFile) map[string][]byte {
	t.Helper()
	body := make(map[string][]byte, len(files))
	for _, file := range files {
		body[file.name] = file.body
	}
	return body
}

func mustStamp(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
