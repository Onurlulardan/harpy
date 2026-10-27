package accesstoken

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/uuid"
)

type tokenFixture struct {
	profile Profile
	private ed25519.PrivateKey
	public  ed25519.PublicKey
	keyID   string
	now     time.Time
	input   ClaimsInput
}

func newTokenFixture(t *testing.T) tokenFixture {
	t.Helper()
	profile, err := NewProfile("http://localhost:8000/auth", "http://localhost:8000/api", true)
	if err != nil {
		t.Fatal(err)
	}
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, ed25519.SeedSize))
	now := time.Unix(2_000_000_000, 0).UTC()
	return tokenFixture{
		profile: profile, private: private, public: private.Public().(ed25519.PublicKey),
		keyID: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{8}, 12)),
		now:   now,
		input: ClaimsInput{
			Subject: uuid.MustParse("10000000-0000-4000-8000-000000000001"),
			Session: uuid.MustParse("20000000-0000-4000-8000-000000000001"),
			Tenant:  uuid.MustParse("30000000-0000-4000-8000-000000000001"),
			Client:  ClientOperatorConsole, IssuedAt: now,
			JTI:          uuid.MustParse("40000000-0000-4000-8000-000000000001"),
			TokenVersion: 3, AuthorizationVersion: 5,
			PlatformAuthorizationClock: 7, SecurityEpoch: [32]byte{1},
			AccessMode: AccessModeOperator, AuthenticationTime: now.Add(-time.Minute),
			Assurance: AssuranceBaseline, Methods: []AuthMethod{AuthMethodPassword},
		},
	}
}

func (fixture tokenFixture) signer(t *testing.T) *Signer {
	t.Helper()
	signer, err := NewSigner(fixture.profile, fixture.keyID, fixture.private)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

func (fixture tokenFixture) verifier(t *testing.T) *Verifier {
	t.Helper()
	verifier, err := NewVerifier(fixture.profile, KeyResolverFunc(
		func(_ context.Context, keyID string) (ed25519.PublicKey, error) {
			if keyID != fixture.keyID {
				return nil, ErrInvalidKey
			}
			return append(ed25519.PublicKey(nil), fixture.public...), nil
		},
	))
	if err != nil {
		t.Fatal(err)
	}
	return verifier
}

func (fixture tokenFixture) claims(t *testing.T) Claims {
	t.Helper()
	claims, err := NewClaims(fixture.profile, fixture.input)
	if err != nil {
		t.Fatal(err)
	}
	return claims
}

func signRaw(private ed25519.PrivateKey, header, payload []byte) string {
	input := encodeBase64URL(header) + "." + encodeBase64URL(payload)
	return input + "." + encodeBase64URL(ed25519.Sign(private, []byte(input)))
}
