package platformiam

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"testing"
	"time"
)

func TestAuthorityAcknowledgementVerifiesExactSignedScope(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	ack := AuthorityAcknowledgement{
		OperationID: fixtureID(1), RegistryBindingID: fixtureID(2), ScopeID: fixtureID(3),
		ScopeKind: EnvironmentScope, RegistrySequence: 4, Generation: 5,
		ProposalSHA256: [32]byte{1}, PreviousHeadSHA256: [32]byte{2},
		StateSHA256: [32]byte{3}, IssuedAt: now,
		RegistryKID: "registry-1",
	}
	ack.AckSHA256 = sha256.Sum256(ack.SigningBytes())
	ack.Signature = ed25519.Sign(privateKey, ack.SigningBytes())
	if err := ack.Verify(now, publicKey); err != nil {
		t.Fatalf("exact signed ack rejected: %v", err)
	}
	ack.Generation++
	if err := ack.Verify(now, publicKey); err == nil {
		t.Fatal("signature verified after generation mutation")
	}
	ack.Generation--
	ack.AckSHA256[0]++
	ack.Signature = ed25519.Sign(privateKey, ack.SigningBytes())
	if err := ack.Verify(now, publicKey); err == nil {
		t.Fatal("acknowledgement accepted a non-canonical artifact digest")
	}
}
