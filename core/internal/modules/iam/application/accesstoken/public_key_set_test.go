package accesstoken

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"testing"
)

func TestPublicKeySetPublishesExactDeterministicJWKS(t *testing.T) {
	firstID := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 12))
	secondID := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{2}, 12))
	firstKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{3}, 32)).Public().(ed25519.PublicKey)
	secondKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{4}, 32)).Public().(ed25519.PublicKey)
	set, err := NewPublicKeySet(map[string]ed25519.PublicKey{
		secondID: secondKey, firstID: firstKey,
	})
	if err != nil {
		t.Fatal(err)
	}
	document, etag, cacheControl, err := set.JWKS()
	if err != nil {
		t.Fatal(err)
	}
	text := string(document)
	if strings.Index(text, firstID) > strings.Index(text, secondID) ||
		strings.Contains(text, `"d"`) || strings.Contains(text, "EdDSA") ||
		strings.Count(text, `"alg":"Ed25519"`) != 2 ||
		!strings.HasPrefix(etag, `"sha256-`) ||
		cacheControl != "public, max-age=300, must-revalidate" {
		t.Fatalf("unexpected JWKS metadata: %s %s %s", text, etag, cacheControl)
	}
	resolved, err := set.AccessTokenKey(context.Background(), firstID)
	if err != nil || !bytes.Equal(resolved, firstKey) {
		t.Fatalf("resolved key mismatch: %x %v", resolved, err)
	}
	resolved[0] ^= 1
	again, _ := set.AccessTokenKey(context.Background(), firstID)
	if bytes.Equal(resolved, again) {
		t.Fatal("public key escaped by reference")
	}
	document[0] ^= 1
	againDocument, _, _, _ := set.JWKS()
	if bytes.Equal(document, againDocument) {
		t.Fatal("JWKS document escaped by reference")
	}
}

func TestPublicKeySetRejectsInvalidAndUnknownKeys(t *testing.T) {
	keyID := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 12))
	tests := []map[string]ed25519.PublicKey{
		nil,
		{"short": bytes.Repeat([]byte{1}, ed25519.PublicKeySize)},
		{keyID: bytes.Repeat([]byte{1}, ed25519.PublicKeySize-1)},
		{keyID: make([]byte, ed25519.PublicKeySize)},
	}
	for _, keys := range tests {
		if _, err := NewPublicKeySet(keys); err != ErrInvalidKey {
			t.Fatalf("invalid key set accepted: %v", err)
		}
	}
	set, err := NewPublicKeySet(map[string]ed25519.PublicKey{
		keyID: bytes.Repeat([]byte{1}, ed25519.PublicKeySize),
	})
	if err != nil {
		t.Fatal(err)
	}
	unknown := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{2}, 12))
	if _, err := set.AccessTokenKey(context.Background(), unknown); err != ErrInvalidKey {
		t.Fatalf("unknown key result: %v", err)
	}
}
