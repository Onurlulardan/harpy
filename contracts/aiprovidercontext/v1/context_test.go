package aiprovidercontextv1

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"
)

func TestSignedContextIsCurrentAndBound(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := NewSigner("core-2026-08", private)
	if err != nil {
		t.Fatal(err)
	}
	defer signer.Close()
	verifier, err := NewVerifier(map[string][]byte{"core-2026-08": public})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	claims := NewClaims(now, 30*time.Second)
	claims.TenantID = "018f936a-0d2d-7d1a-bdf1-9834f1f6a100"
	claims.ActorID = "018f936a-0d2d-7d1a-bdf1-9834f1f6a102"
	claims.ConfigID = "018f936a-0d2d-7d1a-bdf1-9834f1f6a101"
	claims.ConfigRevision, claims.AuthorizationVersion = 7, 11
	claims.Permission, claims.Operation = "ai.config.manage", "ai.config.test"
	claims.CorrelationID = "018f936a-0d2d-7d1a-bdf1-9834f1f6a103"
	claims.JTI = "018f936a-0d2d-7d1a-bdf1-9834f1f6a104"
	token, err := signer.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	expected := Expected{TenantID: claims.TenantID, ConfigID: claims.ConfigID,
		ConfigRevision: 7, AuthorizationVersion: 11, Operation: claims.Operation,
		Now: now.Add(time.Second)}
	if _, err := verifier.Verify(token, expected); err != nil {
		t.Fatal(err)
	}
	expected.ConfigRevision++
	if _, err := verifier.Verify(token, expected); err == nil {
		t.Fatal("stale revision accepted")
	}
	expected.ConfigRevision, expected.Now = 7, now.Add(31*time.Second)
	if _, err := verifier.Verify(token, expected); err == nil {
		t.Fatal("expired context accepted")
	}
}
