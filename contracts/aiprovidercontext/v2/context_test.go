package aiprovidercontextv2

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	aiinferencev1 "github.com/Onurlulardan/harpy/contracts/aiinference/v1"
)

func TestContextBindsEnvelopeInvocationReceiptAndHardDeadline(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, _ := NewSigner("core-2026-08", private)
	defer signer.Close()
	verifier, _ := NewVerifier(map[string][]byte{"core-2026-08": public})
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	claims := NewClaims(now, 30*time.Second)
	claims.TenantID = "018f936a-0d2d-7d1a-bdf1-9834f1f6a100"
	claims.ActorID = "018f936a-0d2d-7d1a-bdf1-9834f1f6a102"
	claims.ConfigID = "018f936a-0d2d-7d1a-bdf1-9834f1f6a101"
	claims.ConfigRevision, claims.CredentialRevision = 7, 9
	claims.AuthorizationVersion = 11
	claims.Permission, claims.Operation = aiinferencev1.PermissionExecute, aiinferencev1.OperationGenerate
	claims.InvocationID = "0198eac5-5ee9-7e0f-8c4a-8493a072f801"
	claims.ReceiptID = "0198eac5-5ee9-7e0f-8c4a-8493a072f802"
	claims.RequestHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	claims.CredentialEnvelopeHash = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	claims.CorrelationID = "018f936a-0d2d-7d1a-bdf1-9834f1f6a103"
	claims.JTI = "018f936a-0d2d-7d1a-bdf1-9834f1f6a104"
	token, err := signer.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	expected := Expected{TenantID: claims.TenantID, ConfigID: claims.ConfigID,
		ConfigRevision: 7, CredentialRevision: 9, AuthorizationVersion: 11,
		Operation: claims.Operation, InvocationID: claims.InvocationID, ReceiptID: claims.ReceiptID,
		RequestHash: claims.RequestHash, CredentialEnvelopeHash: claims.CredentialEnvelopeHash,
		Now: now.Add(time.Second), HardDeadline: now.Add(45 * time.Second)}
	if _, err := verifier.Verify(token, expected); err != nil {
		t.Fatal(err)
	}
	expected.ReceiptID = "0198eac5-5ee9-7e0f-8c4a-8493a072f899"
	if _, err := verifier.Verify(token, expected); err == nil {
		t.Fatal("receipt replay binding changed")
	}
	expected.ReceiptID, expected.HardDeadline = claims.ReceiptID, now.Add(20*time.Second)
	if _, err := verifier.Verify(token, expected); err == nil {
		t.Fatal("context outlived provider hard deadline")
	}
}

func TestVerifierRejectsUnknownSigningKeyWithoutPanicking(t *testing.T) {
	trustedPublic, _, _ := ed25519.GenerateKey(rand.Reader)
	_, untrustedPrivate, _ := ed25519.GenerateKey(rand.Reader)
	signer, _ := NewSigner("unknown-key", untrustedPrivate)
	defer signer.Close()
	verifier, _ := NewVerifier(map[string][]byte{"trusted-key": trustedPublic})
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	claims := NewClaims(now, 10*time.Second)
	claims.TenantID = "018f936a-0d2d-7d1a-bdf1-9834f1f6a100"
	claims.ActorID = "018f936a-0d2d-7d1a-bdf1-9834f1f6a102"
	claims.ConfigID = "018f936a-0d2d-7d1a-bdf1-9834f1f6a101"
	claims.ConfigRevision, claims.CredentialRevision = 7, 0
	claims.AuthorizationVersion = 11
	claims.Permission, claims.Operation = aiinferencev1.PermissionExecute, aiinferencev1.OperationEmbed
	claims.InvocationID = "0198eac5-5ee9-7e0f-8c4a-8493a072f801"
	claims.ReceiptID = "0198eac5-5ee9-7e0f-8c4a-8493a072f802"
	claims.RequestHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	claims.CorrelationID = "018f936a-0d2d-7d1a-bdf1-9834f1f6a103"
	claims.JTI = "018f936a-0d2d-7d1a-bdf1-9834f1f6a104"
	token, err := signer.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	expected := Expected{TenantID: claims.TenantID, ConfigID: claims.ConfigID,
		Operation: claims.Operation, InvocationID: claims.InvocationID, ReceiptID: claims.ReceiptID,
		RequestHash: claims.RequestHash, ConfigRevision: 7, AuthorizationVersion: 11,
		Now: now.Add(time.Second), HardDeadline: now.Add(15 * time.Second)}
	if _, err := verifier.Verify(token, expected); err == nil {
		t.Fatal("unknown signing key accepted")
	}
}
