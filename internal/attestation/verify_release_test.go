package attestation

import (
	"encoding/json"
	"testing"
	"time"
)

func TestVerifyReleaseAcceptsExactSignedContract(t *testing.T) {
	manifest, signature, root := releaseTestDocuments(t)
	verified, err := VerifyRelease(manifest, signature, root, releaseTestNow, "harpy")
	if err != nil {
		t.Fatalf("verify exact release: %v", err)
	}
	if verified.Value().DatabaseOrigin != "v1_adopted" || verified.ManifestSHA256() == "" {
		t.Fatalf("unexpected verified release: %+v", verified)
	}
	if want := time.Date(2026, 7, 13, 11, 59, 0, 0, time.UTC); verified.IssuedAt() != want {
		t.Fatalf("authenticated release issue time = %s, want %s", verified.IssuedAt(), want)
	}
}

func TestVerifyReleaseRejectsTamperAndWrongTarget(t *testing.T) {
	manifest, signature, root := releaseTestDocuments(t)
	if _, err := VerifyRelease(manifest, signature, root, releaseTestNow, "other"); err == nil {
		t.Fatal("wrong target database accepted")
	}
	var tampered map[string]any
	if err := json.Unmarshal(manifest, &tampered); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	tampered["runtime_generation"] = float64(3)
	if _, err := VerifyRelease(mustReleaseJSON(t, tampered), signature, root, releaseTestNow, "harpy"); err == nil {
		t.Fatal("tampered manifest accepted")
	}
}

func TestVerifyReleaseRejectsUnknownAndMissingMembers(t *testing.T) {
	manifest, signature, root := releaseTestDocuments(t)
	var value map[string]any
	if err := json.Unmarshal(manifest, &value); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	value["Database_Name"] = value["database_name"]
	if _, err := VerifyRelease(mustReleaseJSON(t, value), signature, root, releaseTestNow, "harpy"); err == nil {
		t.Fatal("case-folded member accepted")
	}
	delete(value, "Database_Name")
	delete(value, "database_owner")
	if _, err := VerifyRelease(mustReleaseJSON(t, value), signature, root, releaseTestNow, "harpy"); err == nil {
		t.Fatal("missing member accepted")
	}
}
