package iamrestorev1

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func TestCustodyManifestIsCanonicalStrictAndResultBound(t *testing.T) {
	securityEpoch := [32]byte{9}
	rotation := RotationResult{OperationID: "22222222-2222-4222-8222-222222222222",
		Status: "rotated", Generation: 2,
		SecurityEpoch:      base64.RawURLEncoding.EncodeToString(securityEpoch[:]),
		PublicKIDs:         []string{base64.RawURLEncoding.EncodeToString(make([]byte, 16))},
		RevokedGenerations: []int64{1}}
	rotation, err := BindRotationResult(rotation)
	if err != nil {
		t.Fatal(err)
	}
	manifest := CustodyManifest{Schema: CustodyManifestSchema,
		EnvironmentID: "11111111-1111-4111-8111-111111111111",
		Provider:      "test.v1", KeyringHandle: "projects/test/keyrings/iam",
		Rotations: []RotationResult{rotation}}
	document, err := BuildCustodyManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseCustodyManifest(document)
	if err != nil || parsed.Rotations[0].ResultSHA256 != rotation.ResultSHA256 {
		t.Fatalf("parse custody manifest: %#v %v", parsed, err)
	}
	tampered := manifest
	tampered.Rotations = append([]RotationResult(nil), manifest.Rotations...)
	tampered.Rotations[0].Generation++
	if _, err = ParseCustodyManifest(canonicalTestDocument(t, tampered)); err == nil {
		t.Fatal("tampered custody result was accepted")
	}
	invalidEpoch := manifest
	invalidEpoch.Rotations = append([]RotationResult(nil), manifest.Rotations...)
	invalidEpoch.Rotations[0].SecurityEpoch = base64.RawURLEncoding.EncodeToString(make([]byte, 31))
	if _, err = ParseCustodyManifest(canonicalTestDocument(t, invalidEpoch)); err == nil {
		t.Fatal("invalid custody security epoch was accepted")
	}
	invalidEpoch.Rotations[0].SecurityEpoch = base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	if _, err = ParseCustodyManifest(canonicalTestDocument(t, invalidEpoch)); err == nil {
		t.Fatal("zero custody security epoch was accepted")
	}
}

func TestSecurityEvidenceRejectsUnknownOrNonCanonicalInput(t *testing.T) {
	value := validSecurityFixture()
	document := canonicalTestDocument(t, value)
	if _, err := ParseSecurityEvidence(document); err != nil {
		t.Fatal(err)
	}
	unknown := append(document[:len(document)-1], []byte(`,"password":"forbidden"}`)...)
	unknown, _ = canonicaljsonv1.CanonicalizeObject(unknown, MaximumDocumentBytes)
	if _, err := ParseSecurityEvidence(unknown); err == nil {
		t.Fatal("unknown secret field was accepted")
	}
	nonCanonical, _ := json.MarshalIndent(value, "", "  ")
	if _, err := ParseSecurityEvidence(nonCanonical); err == nil {
		t.Fatal("non-canonical evidence was accepted")
	}
}

func validSecurityFixture() SecurityEvidence {
	digest := strings.Repeat("a", 64)
	return SecurityEvidence{Schema: SecurityEvidenceSchema,
		EnvironmentID:          "11111111-1111-4111-8111-111111111111",
		RestoreOperationID:     "22222222-2222-4222-8222-222222222222",
		RecoverySetID:          "33333333-3333-4333-8333-333333333333",
		TenantID:               "44444444-4444-4444-8444-444444444444",
		OwnerUserID:            "55555555-5555-4555-8555-555555555555",
		OwnerBindingID:         "66666666-6666-4666-8666-666666666666",
		OwnerBindingGeneration: 1, AnchorID: "77777777-7777-4777-8777-777777777777",
		NewDatabaseSystemID:  "88888888-8888-4888-8888-888888888888",
		RestoreReceiptSHA256: digest, RecoveryTrustSHA256: digest,
		AckTrustSHA256: digest, CustodyManifestSHA256: digest,
		ProtectedHeadSHA256: digest, ProtectedAckSHA256: digest,
		ProtectedCheckpointSHA256: digest, ProtectedAck: json.RawMessage(`{}`),
		ProtectedCheckpoint: json.RawMessage(`{}`)}
}

func canonicalTestDocument(t *testing.T, value any) []byte {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumDocumentBytes)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}
