package attestation

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func TestSignerAndTrustRootVerifyCanonicalDocument(t *testing.T) {
	root, signer := attestationFixture(t, 1)
	issuedAt := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	signature, err := signer.SignAt([]byte(`{"b":2,"a":1}`), issuedAt)
	if err != nil {
		t.Fatal(err)
	}
	if err := root.Verify([]byte(`{"a":1,"b":2}`), signature, issuedAt); err != nil {
		t.Fatal(err)
	}
	if err := root.Verify([]byte(`{"a":1,"b":3}`), signature, issuedAt); err == nil {
		t.Fatal("tampered document verified")
	}
}

func TestVerificationRejectsWrongIdentityAndValidity(t *testing.T) {
	root, signer := attestationFixture(t, 2)
	signature, err := signer.SignAt([]byte(`{"safe":true}`), fixtureTime())
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name      string
		signature Signature
		issuedAt  time.Time
	}{
		{"unknown key", Signature{algorithmEd25519, "missing", signature.Value}, fixtureTime()},
		{"wrong algorithm", Signature{"Ed25519ph", signature.KeyID, signature.Value}, fixtureTime()},
		{"before validity", signature, time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)},
		{"after validity", signature, time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)},
		{"malformed value", Signature{algorithmEd25519, signature.KeyID, "bad"}, fixtureTime()},
		{"non-canonical base64url", Signature{algorithmEd25519, signature.KeyID, nonCanonicalBase64URL(t, signature.Value)}, fixtureTime()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := root.Verify([]byte(`{"safe":true}`), test.signature, test.issuedAt); err == nil {
				t.Fatal("unsafe signature verified")
			}
		})
	}
}

func nonCanonicalBase64URL(t *testing.T, value string) string {
	t.Helper()
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	index := strings.IndexByte(alphabet, value[len(value)-1])
	if index < 0 || index+1 >= len(alphabet) {
		t.Fatal("test base64url suffix cannot be mutated")
	}
	return value[:len(value)-1] + string(alphabet[index+1])
}

func TestSignerRejectsIssuedAtOutsideKeyValidity(t *testing.T) {
	_, signer := attestationFixture(t, 4)
	for _, issuedAt := range []time.Time{
		time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC),
	} {
		if _, err := signer.SignAt([]byte(`{"safe":true}`), issuedAt); err == nil {
			t.Fatal("signer accepted time outside key validity")
		}
	}
}

func TestTrustAndSigningKeyDocumentsAreStrict(t *testing.T) {
	root, _ := attestationFixture(t, 3)
	wrongSeed := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{4}, ed25519.SeedSize))
	wrongSigner, err := MarshalCanonical(signingKeyDocument{
		ContractVersion: "attestation-signing-key.v1", KeyID: "release-1",
		Algorithm: algorithmEd25519, Seed: wrongSeed,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseSigner(wrongSigner, root); err == nil {
		t.Fatal("untrusted signing key accepted")
	}
	caseFolded := []byte(`{"ContractVersion":"attestation-trust-root.v1","keys":[]}`)
	if _, err := ParseTrustRoot(caseFolded); err == nil {
		t.Fatal("case-folded trust root key accepted")
	}
	duplicate := trustDocumentForSeed(t, 3)
	duplicate.Keys = append(duplicate.Keys, duplicate.Keys[0])
	encoded, err := MarshalCanonical(duplicate)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseTrustRoot(encoded); err == nil {
		t.Fatal("duplicate trust root key ID accepted")
	}
}

func attestationFixture(t *testing.T, seedByte byte) (TrustRoot, Signer) {
	t.Helper()
	trustDocument := trustDocumentForSeed(t, seedByte)
	trustJSON, err := MarshalCanonical(trustDocument)
	if err != nil {
		t.Fatal(err)
	}
	root, err := ParseTrustRoot(trustJSON)
	if err != nil {
		t.Fatal(err)
	}
	seed := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{seedByte}, ed25519.SeedSize))
	signerJSON, err := MarshalCanonical(signingKeyDocument{
		ContractVersion: "attestation-signing-key.v1", KeyID: "release-1",
		Algorithm: algorithmEd25519, Seed: seed,
	})
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ParseSigner(signerJSON, root)
	if err != nil {
		t.Fatal(err)
	}
	return root, signer
}

func trustDocumentForSeed(t *testing.T, seedByte byte) trustRootDocument {
	t.Helper()
	privateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{seedByte}, ed25519.SeedSize))
	publicKey := privateKey.Public().(ed25519.PublicKey)
	return trustRootDocument{
		ContractVersion: "attestation-trust-root.v1",
		Keys: []trustedKeyRecord{{
			KeyID: "release-1", Algorithm: algorithmEd25519,
			PublicKey: base64.RawURLEncoding.EncodeToString(publicKey),
			NotBefore: "2026-07-12T00:00:00Z", NotAfter: "2026-07-14T00:00:00Z",
		}},
	}
}

func fixtureTime() time.Time {
	return time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
}
