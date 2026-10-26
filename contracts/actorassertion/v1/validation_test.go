package actorassertionv1

import (
	"bytes"
	"crypto/ed25519"
	"testing"
)

func TestTargetBindingRejectsAmbiguousInputs(t *testing.T) {
	if _, err := TargetBindingSHA256(); err == nil {
		t.Fatal("empty target binding accepted")
	}
	if _, err := TargetBindingSHA256(""); err == nil {
		t.Fatal("empty target component accepted")
	}
	parts := make([]string, 17)
	for index := range parts {
		parts[index] = "x"
	}
	if _, err := TargetBindingSHA256(parts...); err == nil {
		t.Fatal("oversized target component set accepted")
	}
}

func TestSignerRejectsWrongPrivateKeyAndInvalidDomain(t *testing.T) {
	fixture := newAssertionFixture(t)
	wrong := append(ed25519.PrivateKey(nil), fixture.private...)
	wrong[ed25519.SeedSize] ^= 1
	if _, err := NewProductionSigner("edge-2026-07", wrong); err == nil {
		t.Fatal("inconsistent Ed25519 private key accepted")
	}
	if _, err := NewSigner("bad.key", fixture.private, bytes.NewReader(make([]byte, 16))); err == nil {
		t.Fatal("invalid key ID accepted")
	}
	invalid := fixture.decision
	invalid.TargetDomain = Domain("iam")
	signer, _ := NewSigner("edge-2026-07", fixture.private, bytes.NewReader(make([]byte, 16)))
	defer signer.Close()
	if _, err := signer.Issue(invalid, fixture.now); err == nil {
		t.Fatal("IAM target domain accepted")
	}
}
