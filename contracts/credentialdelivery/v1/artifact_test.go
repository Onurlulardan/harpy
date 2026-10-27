package credentialdeliveryv1

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSealedArtifactRoundTripAndBinding(t *testing.T) {
	operationID := uuid.MustParse("019cba35-a9ad-7ff1-92f3-4c16bd26b60f")
	attemptID := uuid.MustParse("019cba36-2862-7f57-bf3d-71dbdc7337e6")
	credential := []byte(base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32)))
	password := []byte("12345678901234567890123456789012")
	lease := time.Date(2026, 7, 23, 14, 5, 0, 0, time.UTC)
	digest := strings.Repeat("a", 64)

	document, err := SealArtifact(operationID, attemptID, digest, lease, credential, password)
	if err != nil {
		t.Fatal(err)
	}
	artifact, opened, err := OpenArtifact(document, operationID, credential)
	if err != nil {
		t.Fatal(err)
	}
	defer clear(opened)
	if !bytes.Equal(opened, password) || artifact.AttemptID != attemptID.String() ||
		artifact.LeaseExpiresAt != lease.Format(time.RFC3339Nano) {
		t.Fatalf("artifact round trip drifted: %#v", artifact)
	}

	wrong := []byte(base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{8}, 32)))
	if _, _, err = OpenArtifact(document, operationID, wrong); err == nil {
		t.Fatal("artifact opened with a different delivery credential")
	}
}

func TestForgeGuardConformanceVector(t *testing.T) {
	operationID := uuid.MustParse("019cba35-a9ad-7ff1-92f3-4c16bd26b60f")
	attemptID := uuid.MustParse("019cba36-2862-7f57-bf3d-71dbdc7337e6")
	credential := []byte(base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32)))
	document, err := sealArtifactWithNonce(operationID, attemptID, strings.Repeat("a", 64),
		time.Date(2026, 7, 23, 14, 5, 0, 0, time.UTC), credential,
		[]byte("12345678901234567890123456789012"), bytes.Repeat([]byte{1}, 12))
	if err != nil {
		t.Fatal(err)
	}
	const expected = `{"schema":"harpy-installation-credential-delivery-artifact.v1","operation_id":"019cba35-a9ad-7ff1-92f3-4c16bd26b60f","attempt_id":"019cba36-2862-7f57-bf3d-71dbdc7337e6","adapter_id":"forgeguard","claim_token_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","lease_expires_at":"2026-07-23T14:05:00Z","nonce":"AQEBAQEBAQEBAQEB","ciphertext":"2ZEXJLMM_8iuhttMsSWTH0IRaNDHOjJ0eFPmdmhOMJ2kpoT_TMDWezm5fi05oUki"}`
	if string(document) != expected {
		t.Fatalf("ForgeGuard delivery vector drifted: %s", document)
	}
}

func TestParseArtifactRejectsUnknownAndTamperedFields(t *testing.T) {
	operationID := uuid.MustParse("019cba35-a9ad-7ff1-92f3-4c16bd26b60f")
	attemptID := uuid.MustParse("019cba36-2862-7f57-bf3d-71dbdc7337e6")
	credential := []byte(base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32)))
	document, err := SealArtifact(operationID, attemptID, strings.Repeat("b", 64),
		time.Date(2026, 7, 23, 14, 5, 0, 0, time.UTC), credential,
		[]byte("12345678901234567890123456789012"))
	if err != nil {
		t.Fatal(err)
	}
	unknown := append(document[:len(document)-1], []byte(`,"unknown":true}`)...)
	if _, err = ParseArtifact(unknown); err == nil {
		t.Fatal("unknown artifact member accepted")
	}
	tampered := bytes.Replace(document, []byte(attemptID.String()),
		[]byte("019cba36-2862-7f57-bf3d-71dbdc7337e7"), 1)
	if _, _, err = OpenArtifact(tampered, operationID, credential); err == nil {
		t.Fatal("tampered artifact opened")
	}
}
