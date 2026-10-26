package attestation

import (
	"encoding/json"
	"testing"
	"time"
)

func TestVerifyReleaseRejectsExpiredAndExcessiveFutureSkew(t *testing.T) {
	manifest, signature, root := releaseTestDocuments(t)
	if _, err := VerifyRelease(manifest, signature, root, releaseTestNow.Add(11*time.Minute), "harpy"); err == nil {
		t.Fatal("expired release accepted")
	}
	if _, err := VerifyRelease(manifest, signature, root, releaseTestNow.Add(-7*time.Minute), "harpy"); err == nil {
		t.Fatal("release beyond future clock skew accepted")
	}
}

func TestVerifyReleaseRejectsInvalidGenerationAndDatabaseContract(t *testing.T) {
	_, signature, root := releaseTestDocuments(t)
	tests := []struct {
		name  string
		field string
		value any
	}{
		{"zero minimum database generation", "minimum_database_generation", int64(0)},
		{"wrong origin", "database_origin", "legacy"},
		{"wrong locale provider", "database_locale_provider", "icu"},
		{"wrong owner", "database_owner", "postgres"},
		{"checksums disabled", "data_checksums", false},
		{"uppercase digest", "roles_manifest_sha256", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := validReleaseTestManifest()
			value[test.field] = test.value
			manifest := mustReleaseJSON(t, value)
			if _, err := VerifyRelease(manifest, signature, root, releaseTestNow, "harpy"); err == nil {
				t.Fatal("invalid release accepted")
			}
		})
	}
}

func TestVerifyReleaseRejectsTrailingSignatureValue(t *testing.T) {
	manifest, signature, root := releaseTestDocuments(t)
	signature = append(signature, []byte(` {}`)...)
	if _, err := VerifyRelease(manifest, signature, root, releaseTestNow, "harpy"); err == nil {
		t.Fatal("trailing signature value accepted")
	}
}

func TestReleaseManifestJSONShapeIncludesAllFields(t *testing.T) {
	document := mustReleaseJSON(t, validReleaseTestManifest())
	var decoded ReleaseManifest
	if err := json.Unmarshal(document, &decoded); err != nil {
		t.Fatalf("decode release manifest: %v", err)
	}
	if decoded.DatabaseLocaleProvider == "" || decoded.DatabaseOwner == "" {
		t.Fatal("database authority fields missing")
	}
}
