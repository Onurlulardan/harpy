package accesstoken

import (
	"bytes"
	"crypto/ed25519"
	"strings"
	"testing"
)

func TestSignerRejectsInconsistentPrivateKeyAndWipesOnClose(t *testing.T) {
	fixture := newTokenFixture(t)
	bad := append(ed25519.PrivateKey(nil), fixture.private...)
	bad[len(bad)-1] ^= 1
	if _, err := NewSigner(fixture.profile, fixture.keyID, bad); err != ErrInvalidKey {
		t.Fatalf("inconsistent key accepted: %v", err)
	}
	signer := fixture.signer(t)
	signer.Close()
	if _, err := signer.Sign(fixture.claims(t)); err != ErrSignerClosed {
		t.Fatalf("closed signer result: %v", err)
	}
	if signer.key != nil || bytes.Contains(signer.key, []byte{7}) {
		t.Fatal("private key remains reachable")
	}
}

func TestSignerRefusesOversizedWireToken(t *testing.T) {
	fixture := newTokenFixture(t)
	profile, err := NewProfile("https://auth.example/"+strings.Repeat("a", 6000), fixture.profile.Audience(), true)
	if err != nil {
		t.Fatal(err)
	}
	fixture.profile = profile
	signer := fixture.signer(t)
	defer signer.Close()
	if _, err := signer.Sign(fixture.claims(t)); err != ErrInvalidClaims {
		t.Fatalf("expected oversized token rejection, got %v", err)
	}
}
