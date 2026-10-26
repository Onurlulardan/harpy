package cryptography

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestRecoveryCodeParsesWithoutRetainingWireAndRedacts(t *testing.T) {
	secret := bytes.Repeat([]byte{0x5b}, 32)
	encoded := make([]byte, base64.RawURLEncoding.EncodedLen(len(secret)))
	base64.RawURLEncoding.Encode(encoded, secret)
	wire := append([]byte("recovery-v1."), encoded...)
	code, err := NewRecoveryCode(wire)
	if err != nil {
		t.Fatal(err)
	}
	clear(wire)
	clear(encoded)
	defer code.Destroy()
	keyID, err := code.KeyID()
	if err != nil || keyID != "recovery-v1" {
		t.Fatalf("key id=%q err=%v", keyID, err)
	}
	key := bytes.Repeat([]byte{0x6c}, 32)
	got, err := code.HMAC(key)
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte("harpy:iam-recovery-code:v1\x00"))
	_, _ = mac.Write(secret)
	var want [32]byte
	copy(want[:], mac.Sum(nil))
	if got != want {
		t.Fatal("recovery HMAC domain contract drifted")
	}
	assertMaterialRedacted(t, code, secret)
	code.Destroy()
	if _, err := code.HMAC(key); err == nil {
		t.Fatal("destroyed recovery code remained usable")
	}
	clear(secret)
}

func TestRecoveryCodeRejectsMalformedWire(t *testing.T) {
	for _, wire := range [][]byte{
		[]byte("missing-separator"), []byte(".YWJjZA"),
		[]byte("kid.part.extra"), []byte("kid.not_base64!"),
	} {
		if _, err := NewRecoveryCode(wire); err == nil {
			t.Fatalf("malformed recovery code %q was accepted", wire)
		}
	}
}
