package aicredentialv1

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"strings"
	"testing"
)

func fixture(t *testing.T) (*Sealer, *Opener, AAD) {
	t.Helper()
	key, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	sealer, err := NewSealer(key.PublicKey().Bytes())
	if err != nil {
		t.Fatal(err)
	}
	opener, err := NewOpener(key.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	aad := AAD{TenantID: "018f936a-0d2d-7d1a-bdf1-9834f1f6a100",
		ConfigID: "018f936a-0d2d-7d1a-bdf1-9834f1f6a101", ConfigRevision: 7,
		Provider: "openai", Audience: Audience, KeyID: "ai-provider-2026-08"}
	return sealer, opener, aad
}

func TestEnvelopeRoundTripAndRandomness(t *testing.T) {
	sealer, opener, aad := fixture(t)
	secret := []byte("secret-that-must-not-leak")
	first, err := sealer.Seal(secret, aad)
	if err != nil {
		t.Fatal(err)
	}
	second, err := sealer.Seal(secret, aad)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(first, second) {
		t.Fatal("ephemeral key and nonce were reused")
	}
	if strings.Contains(string(first), string(secret)) {
		t.Fatal("plaintext leaked into envelope")
	}
	opened, err := opener.Open(first, aad)
	if err != nil {
		t.Fatal(err)
	}
	defer clear(opened)
	if !bytes.Equal(opened, secret) {
		t.Fatal("opened credential differs")
	}
}

func TestEnvelopeRejectsAADTamperingUnknownAndNonCanonicalJSON(t *testing.T) {
	sealer, opener, aad := fixture(t)
	document, err := sealer.Seal([]byte("secret"), aad)
	if err != nil {
		t.Fatal(err)
	}
	tampered := aad
	tampered.ConfigRevision++
	if value, err := opener.Open(document, tampered); err == nil || value != nil {
		clear(value)
		t.Fatal("AAD tampering was accepted")
	}
	nonCanonical := append([]byte(" "), document...)
	if _, err := Parse(nonCanonical); err == nil {
		t.Fatal("non-canonical envelope accepted")
	}
	withUnknown := append(document[:len(document)-1], []byte(`,"secret":"leak"}`)...)
	if _, err := Parse(withUnknown); err == nil {
		t.Fatal("unknown field accepted")
	}
}

func TestEnvelopeRejectsOversizedPlaintext(t *testing.T) {
	sealer, _, aad := fixture(t)
	secret := bytes.Repeat([]byte{'x'}, MaximumPlaintextBytes+1)
	if document, err := sealer.Seal(secret, aad); err == nil || document != nil {
		t.Fatal("oversized credential accepted")
	}
}
