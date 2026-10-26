package recoverykitv1

import (
	"bytes"
	"crypto/ecdh"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"
)

type browserVectorFixture struct {
	ClaimPrivateBase64URL string          `json:"claim_private_base64url"`
	ClaimPublicBase64URL  string          `json:"claim_public_base64url"`
	ArtifactSHA256        string          `json:"artifact_sha256"`
	Plaintext             json.RawMessage `json:"plaintext"`
	Sealed                json.RawMessage `json:"sealed"`
}

func TestBrowserHPKEVector(t *testing.T) {
	encoded, err := os.ReadFile("testdata/browser-vector.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture browserVectorFixture
	if err := json.Unmarshal(encoded, &fixture); err != nil {
		t.Fatal(err)
	}
	privateBytes, err := base64.RawURLEncoding.DecodeString(fixture.ClaimPrivateBase64URL)
	if err != nil || len(privateBytes) != 32 {
		t.Fatal("browser vector claim private key is invalid")
	}
	var private [32]byte
	copy(private[:], privateBytes)
	key, err := ecdh.X25519().NewPrivateKey(private[:])
	if err != nil || base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()) !=
		fixture.ClaimPublicBase64URL {
		t.Fatal("browser vector claim key pair does not match")
	}
	sealed := compactJSON(t, fixture.Sealed)
	if digestHex(sealed) != fixture.ArtifactSHA256 {
		t.Fatal("browser vector sealed artifact digest does not match")
	}
	opened, err := Open(sealed, private, OpenExpected{
		EnvironmentID: kitEnvironmentID, OperationID: kitOperationID,
		Generation: 1, Now: time.Date(2026, 7, 23, 6, 1, 0, 0, time.UTC),
	})
	if err != nil || !bytes.Equal(opened.CanonicalBytes(), compactJSON(t, fixture.Plaintext)) {
		t.Fatal("browser vector did not open to the fixed plaintext")
	}
}

func compactJSON(t *testing.T, value []byte) []byte {
	t.Helper()
	var result bytes.Buffer
	if err := json.Compact(&result, value); err != nil {
		t.Fatal(err)
	}
	return result.Bytes()
}
